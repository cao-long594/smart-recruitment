package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/cloudwego/eino/schema"
	// DashScope 兼容 HTTP 对话接口；依赖 eino-ext 下第三方包路径（与业务侧仅使用 Qwen 配置无关）。
	chatmodel "github.com/cloudwego/eino-ext/components/model/openai"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	recruitmentv1 "recruitment/api/gen/go/recruitment/v1"
	"recruitment/logic-grpc-service/internal/config"
	"recruitment/logic-grpc-service/internal/meta"
	"recruitment/logic-grpc-service/internal/model"
	"recruitment/logic-grpc-service/internal/osssvc"
)

type Recruitment struct {
	recruitmentv1.UnimplementedRecruitmentServiceServer
	DB     *gorm.DB
	OSS    *osssvc.Client
	Config *config.Config
	chat   *chatmodel.ChatModel
}

func NewRecruitment(db *gorm.DB, oss *osssvc.Client, cfg *config.Config) (*Recruitment, error) {
	r := &Recruitment{DB: db, OSS: oss, Config: cfg}
	if cfg.QwenAPIKey != "" {
		cm, err := chatmodel.NewChatModel(context.Background(), &chatmodel.ChatModelConfig{
			APIKey:  cfg.QwenAPIKey,
			Model:   cfg.QwenModel,
			BaseURL: cfg.QwenBaseURL,
		})
		if err != nil {
			return nil, fmt.Errorf("eino chat model: %w", err)
		}
		r.chat = cm
	}
	return r, nil
}

func roleToDB(r recruitmentv1.Role) (string, error) {
	switch r {
	case recruitmentv1.Role_ROLE_HR:
		return "hr", nil
	case recruitmentv1.Role_ROLE_CANDIDATE:
		return "candidate", nil
	default:
		return "", status.Error(codes.InvalidArgument, "invalid role")
	}
}

func roleToProto(s string) recruitmentv1.Role {
	switch s {
	case "hr":
		return recruitmentv1.Role_ROLE_HR
	case "candidate":
		return recruitmentv1.Role_ROLE_CANDIDATE
	default:
		return recruitmentv1.Role_ROLE_UNSPECIFIED
	}
}

func (s *Recruitment) Register(ctx context.Context, in *recruitmentv1.RegisterRequest) (*recruitmentv1.RegisterResponse, error) {
	role, err := roleToDB(in.GetRole())
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	if email == "" || in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "hash failed")
	}
	u := model.User{Email: email, PasswordHash: string(hash), Role: role}
	if err := s.DB.Create(&u).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, status.Error(codes.AlreadyExists, "email exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.RegisterResponse{UserId: u.ID}, nil
}

func (s *Recruitment) Login(ctx context.Context, in *recruitmentv1.LoginRequest) (*recruitmentv1.LoginResponse, error) {
	email := strings.TrimSpace(strings.ToLower(in.GetEmail()))
	var u model.User
	if err := s.DB.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.GetPassword())) != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return &recruitmentv1.LoginResponse{UserId: u.ID, Role: roleToProto(u.Role)}, nil
}

func jobToProto(j *model.Job) *recruitmentv1.Job {
	return &recruitmentv1.Job{
		Id:          j.ID,
		HrUserId:    j.HRUserID,
		Title:       j.Title,
		Description: j.Description,
		Status:      j.Status,
		CreatedAt:   j.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   j.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *Recruitment) ListPublicJobs(ctx context.Context, in *recruitmentv1.ListPublicJobsRequest) (*recruitmentv1.ListPublicJobsResponse, error) {
	page := int(in.GetPage())
	if page < 1 {
		page = 1
	}
	ps := int(in.GetPageSize())
	if ps < 1 || ps > 100 {
		ps = 20
	}
	var total int64
	q := s.DB.Model(&model.Job{}).Where("status = ?", "active")
	if err := q.Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var jobs []model.Job
	if err := q.Order("id desc").Offset((page - 1) * ps).Limit(ps).Find(&jobs).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*recruitmentv1.Job, 0, len(jobs))
	for i := range jobs {
		out = append(out, jobToProto(&jobs[i]))
	}
	return &recruitmentv1.ListPublicJobsResponse{Jobs: out, Total: total}, nil
}

func (s *Recruitment) ListMyJobs(ctx context.Context, in *recruitmentv1.ListMyJobsRequest) (*recruitmentv1.ListMyJobsResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustHR(u); err != nil {
		return nil, err
	}
	page := int(in.GetPage())
	if page < 1 {
		page = 1
	}
	ps := int(in.GetPageSize())
	if ps < 1 || ps > 100 {
		ps = 20
	}
	var total int64
	q := s.DB.Model(&model.Job{}).Where("hr_user_id = ?", u.ID)
	if err := q.Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var jobs []model.Job
	if err := q.Order("id desc").Offset((page - 1) * ps).Limit(ps).Find(&jobs).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*recruitmentv1.Job, 0, len(jobs))
	for i := range jobs {
		out = append(out, jobToProto(&jobs[i]))
	}
	return &recruitmentv1.ListMyJobsResponse{Jobs: out, Total: total}, nil
}

func (s *Recruitment) CreateJob(ctx context.Context, in *recruitmentv1.CreateJobRequest) (*recruitmentv1.CreateJobResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustHR(u); err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.GetTitle())
	if title == "" {
		return nil, status.Error(codes.InvalidArgument, "title required")
	}
	j := model.Job{HRUserID: u.ID, Title: title, Description: in.GetDescription(), Status: "active"}
	if err := s.DB.Create(&j).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.CreateJobResponse{Job: jobToProto(&j)}, nil
}

func (s *Recruitment) UpdateJob(ctx context.Context, in *recruitmentv1.UpdateJobRequest) (*recruitmentv1.UpdateJobResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustHR(u); err != nil {
		return nil, err
	}
	var j model.Job
	if err := s.DB.First(&j, in.GetJobId()).Error; err != nil {
		return nil, status.Error(codes.NotFound, "job not found")
	}
	if j.HRUserID != u.ID {
		return nil, status.Error(codes.PermissionDenied, "not owner")
	}
	if in.GetTitle() != "" {
		j.Title = strings.TrimSpace(in.GetTitle())
	}
	if in.GetDescription() != "" {
		j.Description = in.GetDescription()
	}
	if err := s.DB.Save(&j).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.UpdateJobResponse{Job: jobToProto(&j)}, nil
}

func (s *Recruitment) ArchiveJob(ctx context.Context, in *recruitmentv1.ArchiveJobRequest) (*recruitmentv1.ArchiveJobResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustHR(u); err != nil {
		return nil, err
	}
	var j model.Job
	if err := s.DB.First(&j, in.GetJobId()).Error; err != nil {
		return nil, status.Error(codes.NotFound, "job not found")
	}
	if j.HRUserID != u.ID {
		return nil, status.Error(codes.PermissionDenied, "not owner")
	}
	if j.Status == "active" {
		j.Status = "archived"
	} else {
		j.Status = "active"
	}
	if err := s.DB.Save(&j).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.ArchiveJobResponse{}, nil
}

func profileComplete(p *model.CandidateProfile) bool {
	if p == nil {
		return false
	}
	nonEmpty := func(s string) bool {
		s = strings.TrimSpace(s)
		if s == "" {
			return false
		}
		for _, r := range s {
			if unicode.IsGraphic(r) {
				return true
			}
		}
		return false
	}
	return nonEmpty(p.Name) && nonEmpty(p.Phone) && nonEmpty(p.Education) && nonEmpty(p.School) &&
		nonEmpty(p.Experience) && nonEmpty(p.Skills)
}

func profileToProto(p *model.CandidateProfile, r *model.Resume) *recruitmentv1.CandidateProfile {
	if p == nil {
		return &recruitmentv1.CandidateProfile{}
	}
	cp := &recruitmentv1.CandidateProfile{
		UserId:          p.UserID,
		Name:            p.Name,
		Phone:           p.Phone,
		Education:       p.Education,
		School:          p.School,
		Experience:      p.Experience,
		Skills:          p.Skills,
		ProfileComplete: profileComplete(p),
	}
	if r != nil && r.ObjectKey != "" {
		cp.HasResume = true
		cp.ResumeObjectKey = r.ObjectKey
		cp.ResumeFileName = r.FileName
	}
	return cp
}

func (s *Recruitment) GetCandidateProfile(ctx context.Context, in *recruitmentv1.GetCandidateProfileRequest) (*recruitmentv1.GetCandidateProfileResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustCandidate(u); err != nil {
		return nil, err
	}
	var p model.CandidateProfile
	_ = s.DB.Where("user_id = ?", u.ID).First(&p).Error
	if p.UserID == 0 {
		p.UserID = u.ID
	}
	var r model.Resume
	_ = s.DB.Where("user_id = ?", u.ID).First(&r).Error
	return &recruitmentv1.GetCandidateProfileResponse{Profile: profileToProto(&p, &r)}, nil
}

func (s *Recruitment) UpdateCandidateProfile(ctx context.Context, in *recruitmentv1.UpdateCandidateProfileRequest) (*recruitmentv1.UpdateCandidateProfileResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustCandidate(u); err != nil {
		return nil, err
	}
	var p model.CandidateProfile
	err = s.DB.Where("user_id = ?", u.ID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		p = model.CandidateProfile{UserID: u.ID}
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	p.Name = in.GetName()
	p.Phone = in.GetPhone()
	p.Education = in.GetEducation()
	p.School = in.GetSchool()
	p.Experience = in.GetExperience()
	p.Skills = in.GetSkills()
	if err := s.DB.Save(&p).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var r model.Resume
	_ = s.DB.Where("user_id = ?", u.ID).First(&r).Error
	return &recruitmentv1.UpdateCandidateProfileResponse{Profile: profileToProto(&p, &r)}, nil
}

var allowedResumeCT = map[string]struct{}{
	"application/pdf":    {},
	"application/msword": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
}

func extAllowed(ext string) bool {
	switch strings.ToLower(ext) {
	case ".pdf", ".doc", ".docx":
		return true
	default:
		return false
	}
}

func validateResumeMagic(head []byte, ext string) error {
	if len(head) < 4 {
		return fmt.Errorf("file too small")
	}
	switch strings.ToLower(ext) {
	case ".pdf":
		if string(head[0:4]) != "%PDF" {
			return fmt.Errorf("not a pdf")
		}
	case ".doc":
		if head[0] != 0xD0 || head[1] != 0xCF || head[2] != 0x11 || head[3] != 0xE0 {
			return fmt.Errorf("not a doc")
		}
	case ".docx":
		if head[0] != 0x50 || head[1] != 0x4B {
			return fmt.Errorf("not a docx")
		}
	default:
		return fmt.Errorf("bad ext")
	}
	return nil
}

func (s *Recruitment) GetResumeUploadCredential(ctx context.Context, in *recruitmentv1.GetResumeUploadCredentialRequest) (*recruitmentv1.GetResumeUploadCredentialResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustCandidate(u); err != nil {
		return nil, err
	}
	if s.OSS == nil {
		return nil, status.Error(codes.FailedPrecondition, "oss not configured")
	}
	fn := in.GetFileName()
	ext := filepath.Ext(fn)
	if !extAllowed(ext) {
		return nil, status.Error(codes.InvalidArgument, "only pdf doc docx")
	}
	ct := strings.TrimSpace(in.GetContentType())
	if _, ok := allowedResumeCT[ct]; !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid content type")
	}
	key := osssvc.SafeResumeObjectKey(u.ID, fn)
	url, hdrs, err := s.OSS.PresignPut(ctx, key, ct, 20*time.Minute)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.GetResumeUploadCredentialResponse{
		UploadUrl: url,
		ObjectKey: key,
		Headers:   hdrs,
	}, nil
}

func (s *Recruitment) ConfirmResumeUploaded(ctx context.Context, in *recruitmentv1.ConfirmResumeUploadedRequest) (*recruitmentv1.ConfirmResumeUploadedResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustCandidate(u); err != nil {
		return nil, err
	}
	if s.OSS == nil {
		return nil, status.Error(codes.FailedPrecondition, "oss not configured")
	}
	key := strings.TrimSpace(in.GetObjectKey())
	if !strings.HasPrefix(key, fmt.Sprintf("resumes/%d/", u.ID)) {
		return nil, status.Error(codes.PermissionDenied, "invalid object key")
	}
	ext := filepath.Ext(in.GetFileName())
	if !extAllowed(ext) {
		return nil, status.Error(codes.InvalidArgument, "bad file type")
	}
	ct := strings.TrimSpace(in.GetContentType())
	if _, ok := allowedResumeCT[ct]; !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid content type")
	}
	head, err := s.OSS.GetObjectHead(ctx, key, 16)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "cannot verify upload: "+err.Error())
	}
	if err := validateResumeMagic(head, ext); err != nil {
		return nil, status.Error(codes.InvalidArgument, "file header mismatch: "+err.Error())
	}
	r := model.Resume{
		UserID:      u.ID,
		ObjectKey:   key,
		FileName:    filepath.Base(in.GetFileName()),
		ContentType: ct,
		SizeBytes:   in.GetSizeBytes(),
	}
	if err := s.DB.Save(&r).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.ConfirmResumeUploadedResponse{}, nil
}

type resumeExtract struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Education  string `json:"education"`
	School     string `json:"school"`
	Experience string `json:"experience"`
	Skills     string `json:"skills"`
}

func extractDOCXText(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var doc *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			doc = f
			break
		}
	}
	if doc == nil {
		return "", fmt.Errorf("word/document.xml not found")
	}
	rc, err := doc.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	dec := xml.NewDecoder(rc)
	var b strings.Builder
	inText := false
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "tab":
				b.WriteByte('\t')
			case "br", "p":
				if b.Len() > 0 {
					b.WriteByte('\n')
				}
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
				b.WriteByte(' ')
			}
		case xml.CharData:
			if inText {
				b.Write([]byte(t))
			}
		}
	}
	return normalizeResumeText(b.String()), nil
}

func normalizeResumeText(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			out = append(out, line)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func parseResumeJSON(raw string) (resumeExtract, bool) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	start := strings.IndexByte(raw, '{')
	end := strings.LastIndexByte(raw, '}')
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}
	var x resumeExtract
	if err := json.Unmarshal([]byte(raw), &x); err != nil {
		return resumeExtract{}, false
	}
	return x, true
}

func firstMatch(text string, re *regexp.Regexp) string {
	m := re.FindStringSubmatch(text)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if len(m) == 1 {
		return strings.TrimSpace(m[0])
	}
	return ""
}

func extractResumeByRules(text string) resumeExtract {
	lines := strings.Split(text, "\n")
	x := resumeExtract{
		Phone: firstMatch(text, regexp.MustCompile(`(?:\+?86[- ]?)?1[3-9]\d{9}`)),
	}
	for _, line := range lines {
		l := strings.TrimSpace(line)
		compact := strings.ReplaceAll(l, " ", "")
		if x.Name == "" {
			if v := firstMatch(compact, regexp.MustCompile(`姓名[:：]?([\p{Han}A-Za-z·]{2,24})`)); v != "" {
				x.Name = v
			}
		}
		if x.School == "" {
			if v := firstMatch(l, regexp.MustCompile(`(?:毕业院校|学校|院校)[:： ]*([^,，;；\n]{2,40})`)); v != "" {
				x.School = v
			} else if strings.Contains(l, "大学") || strings.Contains(l, "学院") {
				x.School = l
			}
		}
		if x.Education == "" {
			for _, edu := range []string{"博士", "硕士", "研究生", "本科", "大专", "专科", "高中"} {
				if strings.Contains(l, edu) {
					x.Education = edu
					break
				}
			}
		}
	}
	if x.Name == "" && len(lines) > 0 {
		candidate := strings.TrimSpace(lines[0])
		if len([]rune(candidate)) <= 24 && !strings.Contains(candidate, "简历") {
			x.Name = candidate
		}
	}
	skillWords := []string{
		"Go", "Golang", "Java", "Python", "JavaScript", "TypeScript", "React", "Vue", "MySQL",
		"Redis", "Docker", "Kubernetes", "Linux", "Gin", "gRPC", "HTML", "CSS", "Android", "AI",
	}
	seen := map[string]bool{}
	var skills []string
	lower := strings.ToLower(text)
	for _, w := range skillWords {
		if strings.Contains(lower, strings.ToLower(w)) && !seen[strings.ToLower(w)] {
			seen[strings.ToLower(w)] = true
			skills = append(skills, w)
		}
	}
	x.Skills = strings.Join(skills, ",")
	if len([]rune(text)) > 1200 {
		x.Experience = string([]rune(text)[:1200])
	} else {
		x.Experience = text
	}
	return x
}

func mergeExtract(base, fallback resumeExtract) resumeExtract {
	if strings.TrimSpace(base.Name) == "" {
		base.Name = fallback.Name
	}
	if strings.TrimSpace(base.Phone) == "" {
		base.Phone = fallback.Phone
	}
	if strings.TrimSpace(base.Education) == "" {
		base.Education = fallback.Education
	}
	if strings.TrimSpace(base.School) == "" {
		base.School = fallback.School
	}
	if strings.TrimSpace(base.Experience) == "" {
		base.Experience = fallback.Experience
	}
	if strings.TrimSpace(base.Skills) == "" {
		base.Skills = fallback.Skills
	}
	return base
}

func (s *Recruitment) extractResumeProfile(ctx context.Context, text string) (resumeExtract, error) {
	fallback := extractResumeByRules(text)
	if s.chat == nil {
		return fallback, nil
	}
	if len([]rune(text)) > 6000 {
		text = string([]rune(text)[:6000])
	}
	msgs := []*schema.Message{
		schema.SystemMessage("你是招聘系统的简历结构化抽取器。只返回一个 JSON 对象，不要 Markdown，不要解释。字段必须是 name、phone、education、school、experience、skills。缺失字段返回空字符串。skills 用英文逗号分隔。"),
		schema.UserMessage("请从下面简历文本抽取候选人档案信息：\n" + text),
	}
	resp, err := s.chat.Generate(ctx, msgs)
	if err != nil {
		return fallback, nil
	}
	ai, ok := parseResumeJSON(resp.Content)
	if !ok {
		return fallback, nil
	}
	return mergeExtract(ai, fallback), nil
}

func (s *Recruitment) ParseResumeToProfile(ctx context.Context, in *recruitmentv1.ParseResumeToProfileRequest) (*recruitmentv1.ParseResumeToProfileResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustCandidate(u); err != nil {
		return nil, err
	}
	var r model.Resume
	_ = s.DB.Where("user_id = ?", u.ID).First(&r).Error
	fileName := r.FileName
	data := in.GetFileContent()
	if len(data) == 0 && s.OSS == nil {
		return nil, status.Error(codes.FailedPrecondition, "oss not configured")
	}
	if len(data) > 0 {
		fileName = filepath.Base(in.GetFileName())
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, status.Error(codes.FailedPrecondition, "upload resume first")
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".docx" {
		return nil, status.Error(codes.FailedPrecondition, "auto parse currently supports docx only")
	}
	const maxResumeBytes = 8 << 20
	if len(data) > maxResumeBytes {
		return nil, status.Error(codes.InvalidArgument, "resume file too large")
	}
	if len(data) == 0 {
		if r.ObjectKey == "" {
			return nil, status.Error(codes.FailedPrecondition, "upload resume first")
		}
		data, err = s.OSS.GetObjectBytes(ctx, r.ObjectKey, maxResumeBytes)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, "cannot read resume: "+err.Error())
		}
		if len(data) > maxResumeBytes {
			return nil, status.Error(codes.InvalidArgument, "resume file too large")
		}
	}
	if err := validateResumeMagic(data, ext); err != nil {
		return nil, status.Error(codes.InvalidArgument, "file header mismatch: "+err.Error())
	}
	text, err := extractDOCXText(data)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot parse docx: "+err.Error())
	}
	if strings.TrimSpace(text) == "" {
		return nil, status.Error(codes.InvalidArgument, "resume has no readable text")
	}
	x, err := s.extractResumeProfile(ctx, text)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	p := &model.CandidateProfile{
		UserID:     u.ID,
		Name:       strings.TrimSpace(x.Name),
		Phone:      strings.TrimSpace(x.Phone),
		Education:  strings.TrimSpace(x.Education),
		School:     strings.TrimSpace(x.School),
		Experience: strings.TrimSpace(x.Experience),
		Skills:     strings.TrimSpace(x.Skills),
	}
	return &recruitmentv1.ParseResumeToProfileResponse{Profile: profileToProto(p, &r)}, nil
}

func (s *Recruitment) GetResumeDownloadURL(ctx context.Context, in *recruitmentv1.GetResumeDownloadURLRequest) (*recruitmentv1.GetResumeDownloadURLResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustHR(u); err != nil {
		return nil, err
	}
	if s.OSS == nil {
		return nil, status.Error(codes.FailedPrecondition, "oss not configured")
	}
	candID := in.GetCandidateUserId()
	var n int64
	if err := s.DB.Model(&model.Application{}).
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.candidate_user_id = ? AND jobs.hr_user_id = ?", candID, u.ID).
		Count(&n).Error; err != nil || n == 0 {
		return nil, status.Error(codes.PermissionDenied, "no application to your jobs")
	}
	var r model.Resume
	if err := s.DB.Where("user_id = ?", candID).First(&r).Error; err != nil {
		return nil, status.Error(codes.NotFound, "resume not found")
	}
	url, err := s.OSS.PresignGet(ctx, r.ObjectKey, 15*time.Minute)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.GetResumeDownloadURLResponse{DownloadUrl: url}, nil
}

func (s *Recruitment) ApplyJob(ctx context.Context, in *recruitmentv1.ApplyJobRequest) (*recruitmentv1.ApplyJobResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustCandidate(u); err != nil {
		return nil, err
	}
	var p model.CandidateProfile
	if err := s.DB.Where("user_id = ?", u.ID).First(&p).Error; err != nil {
		return nil, status.Error(codes.FailedPrecondition, "complete profile first")
	}
	if !profileComplete(&p) {
		return nil, status.Error(codes.FailedPrecondition, "profile incomplete")
	}
	var r model.Resume
	if err := s.DB.Where("user_id = ?", u.ID).First(&r).Error; err != nil || r.ObjectKey == "" {
		return nil, status.Error(codes.FailedPrecondition, "upload resume first")
	}
	var j model.Job
	if err := s.DB.First(&j, in.GetJobId()).Error; err != nil {
		return nil, status.Error(codes.NotFound, "job not found")
	}
	if j.Status != "active" {
		return nil, status.Error(codes.FailedPrecondition, "job not active")
	}
	var exists int64
	if err := s.DB.Model(&model.Application{}).Where("job_id = ? AND candidate_user_id = ?", j.ID, u.ID).Count(&exists).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exists > 0 {
		return nil, status.Error(codes.AlreadyExists, "already applied")
	}
	app := model.Application{JobID: j.ID, CandidateUserID: u.ID}
	if err := s.DB.Create(&app).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.ApplyJobResponse{}, nil
}

func (s *Recruitment) ListApplicationsForJob(ctx context.Context, in *recruitmentv1.ListApplicationsForJobRequest) (*recruitmentv1.ListApplicationsForJobResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustHR(u); err != nil {
		return nil, err
	}
	var j model.Job
	if err := s.DB.First(&j, in.GetJobId()).Error; err != nil {
		return nil, status.Error(codes.NotFound, "job not found")
	}
	if j.HRUserID != u.ID {
		return nil, status.Error(codes.PermissionDenied, "not owner")
	}
	page := int(in.GetPage())
	if page < 1 {
		page = 1
	}
	ps := int(in.GetPageSize())
	if ps < 1 || ps > 100 {
		ps = 20
	}
	var total int64
	if err := s.DB.Model(&model.Application{}).Where("job_id = ?", j.ID).Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var apps []model.Application
	if err := s.DB.Where("job_id = ?", j.ID).Order("id desc").Offset((page - 1) * ps).Limit(ps).Find(&apps).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	rows := make([]*recruitmentv1.ApplicationRow, 0, len(apps))
	for _, a := range apps {
		var p model.CandidateProfile
		_ = s.DB.Where("user_id = ?", a.CandidateUserID).First(&p).Error
		var r model.Resume
		_ = s.DB.Where("user_id = ?", a.CandidateUserID).First(&r).Error
		rows = append(rows, &recruitmentv1.ApplicationRow{
			Id:              a.ID,
			JobId:           a.JobID,
			CandidateUserId: a.CandidateUserID,
			AppliedAt:       a.CreatedAt.Format(time.RFC3339),
			Candidate:       profileToProto(&p, &r),
		})
	}
	return &recruitmentv1.ListApplicationsForJobResponse{Rows: rows, Total: total}, nil
}

func (s *Recruitment) analyticsFacts(ctx context.Context, question string) (string, error) {
	q := strings.TrimSpace(question)
	ql := strings.ToLower(q)
	var b strings.Builder

	var totalApps int64
	if err := s.DB.Model(&model.Application{}).Count(&totalApps).Error; err != nil {
		return "", err
	}
	if strings.Contains(q, "投递") && (strings.Contains(q, "总数") || strings.Contains(q, "多少") || strings.Contains(ql, "total")) {
		fmt.Fprintf(&b, "全平台投递记录条数: %d\n", totalApps)
	}

	var activeJobs int64
	_ = s.DB.Model(&model.Job{}).Where("status = ?", "active").Count(&activeJobs).Error
	if strings.Contains(q, "岗位") && (strings.Contains(q, "活跃") || strings.Contains(q, "在招") || strings.Contains(ql, "active")) {
		fmt.Fprintf(&b, "在招岗位数(status=active): %d\n", activeJobs)
	}

	// hot jobs: top 5 by application count
	if strings.Contains(q, "热度") || strings.Contains(q, "热门") || strings.Contains(ql, "hot") {
		type row struct {
			JobID int64
			Cnt   int64
		}
		var rs []row
		_ = s.DB.Model(&model.Application{}).
			Select("job_id, COUNT(*) as cnt").
			Group("job_id").
			Order("cnt desc").
			Limit(5).
			Scan(&rs).Error
		for _, r := range rs {
			var j model.Job
			_ = s.DB.First(&j, r.JobID).Error
			fmt.Fprintf(&b, "岗位 #%d %s 投递数: %d\n", j.ID, j.Title, r.Cnt)
		}
	}

	if b.Len() == 0 {
		fmt.Fprintf(&b, "全平台投递记录条数: %d\n", totalApps)
		fmt.Fprintf(&b, "在招岗位数: %d\n", activeJobs)
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *Recruitment) ChatMessage(ctx context.Context, in *recruitmentv1.ChatMessageRequest) (*recruitmentv1.ChatMessageResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustHR(u); err != nil {
		return nil, err
	}
	q := strings.TrimSpace(in.GetContent())
	if q == "" {
		return nil, status.Error(codes.InvalidArgument, "empty message")
	}
	um := model.ChatMessage{HRUserID: u.ID, Role: "user", Content: q}
	if err := s.DB.Create(&um).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	facts, err := s.analyticsFacts(ctx, q)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var answer string
	if s.chat != nil {
		msgs := []*schema.Message{
			schema.SystemMessage("你是招聘系统数据分析助手。只根据「业务事实」回答，不要编造数字。若事实不足请说明。"),
			schema.UserMessage("业务事实：\n" + facts + "\n\n用户问题：\n" + q),
		}
		resp, err := s.chat.Generate(ctx, msgs)
		if err != nil {
			return nil, status.Error(codes.Internal, "eino: "+err.Error())
		}
		answer = resp.Content
	} else {
		answer = "（未配置通义千问 API Key，请在 config 的 qwen 段或环境变量 QWEN_API_KEY 中配置；以下为数据库汇总事实）\n" + facts
	}
	am := model.ChatMessage{HRUserID: u.ID, Role: "assistant", Content: answer}
	if err := s.DB.Create(&am).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &recruitmentv1.ChatMessageResponse{Answer: answer}, nil
}

func (s *Recruitment) ListChatHistory(ctx context.Context, in *recruitmentv1.ListChatHistoryRequest) (*recruitmentv1.ListChatHistoryResponse, error) {
	u, err := meta.IncomingUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := meta.MustHR(u); err != nil {
		return nil, err
	}
	page := int(in.GetPage())
	if page < 1 {
		page = 1
	}
	ps := int(in.GetPageSize())
	if ps < 1 || ps > 200 {
		ps = 100
	}
	var total int64
	if err := s.DB.Model(&model.ChatMessage{}).Where("hr_user_id = ?", u.ID).Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var msgs []model.ChatMessage
	if err := s.DB.Where("hr_user_id = ?", u.ID).Order("id asc").Offset((page - 1) * ps).Limit(ps).Find(&msgs).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*recruitmentv1.ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, &recruitmentv1.ChatMessage{
			Id:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}
	return &recruitmentv1.ListChatHistoryResponse{Messages: out, Total: total}, nil
}
