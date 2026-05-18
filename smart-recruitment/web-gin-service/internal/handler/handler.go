package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	recruitmentv1 "recruitment/api/gen/go/recruitment/v1"
	"recruitment/web-gin-service/internal/config"
	"recruitment/web-gin-service/internal/jwtutil"
)

const (
	metaUserID   = "x-user-id"
	metaUserRole = "x-user-role"
)

type Handler struct {
	Cfg    *config.Config
	Client recruitmentv1.RecruitmentServiceClient
}

func grpcMetaCtx(c *gin.Context, userID int64, role string) context.Context {
	md := metadata.Pairs(metaUserID, strconv.FormatInt(userID, 10), metaUserRole, role)
	return metadata.NewOutgoingContext(c.Request.Context(), md)
}

func (h *Handler) grpcErr(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	switch st.Code() {
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
	case codes.InvalidArgument, codes.FailedPrecondition:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
	}
}

func roleProto(r string) recruitmentv1.Role {
	switch r {
	case "hr":
		return recruitmentv1.Role_ROLE_HR
	case "candidate":
		return recruitmentv1.Role_ROLE_CANDIDATE
	default:
		return recruitmentv1.Role_ROLE_UNSPECIFIED
	}
}

func (h *Handler) Register(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	resp, err := h.Client.Register(c.Request.Context(), &recruitmentv1.RegisterRequest{
		Email:    body.Email,
		Password: body.Password,
		Role:     roleProto(strings.ToLower(body.Role)),
	})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": resp.UserId})
}

func (h *Handler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	resp, err := h.Client.Login(c.Request.Context(), &recruitmentv1.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	r := ""
	switch resp.Role {
	case recruitmentv1.Role_ROLE_HR:
		r = "hr"
	case recruitmentv1.Role_ROLE_CANDIDATE:
		r = "candidate"
	}
	token, err := jwtutil.Sign(h.Cfg.JWTSecret, resp.UserId, r, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": resp.UserId, "role": r})
}

func authID(c *gin.Context) (string, bool) {
	h := c.GetHeader("Authorization")
	const p = "Bearer "
	if len(h) < len(p) || !strings.EqualFold(h[:len(p)], p) {
		return "", false
	}
	return strings.TrimSpace(h[len(p):]), true
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok, ok := authID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		claims, err := jwtutil.Parse(h.Cfg.JWTSecret, tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (h *Handler) ListPublicJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	resp, err := h.Client.ListPublicJobs(c.Request.Context(), &recruitmentv1.ListPublicJobsRequest{
		Page: int32(page), PageSize: int32(ps),
	})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListMyJobs(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	resp, err := h.Client.ListMyJobs(ctx, &recruitmentv1.ListMyJobsRequest{Page: int32(page), PageSize: int32(ps)})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateJob(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	_ = c.ShouldBindJSON(&body)
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.CreateJob(ctx, &recruitmentv1.CreateJobRequest{Title: body.Title, Description: body.Description})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateJob(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	_ = c.ShouldBindJSON(&body)
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.UpdateJob(ctx, &recruitmentv1.UpdateJobRequest{JobId: id, Title: body.Title, Description: body.Description})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ArchiveJob(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	_, err := h.Client.ArchiveJob(ctx, &recruitmentv1.ArchiveJobRequest{JobId: id})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) GetProfile(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.GetCandidateProfile(ctx, &recruitmentv1.GetCandidateProfileRequest{})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	var body recruitmentv1.UpdateCandidateProfileRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.UpdateCandidateProfile(ctx, &body)
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ResumeUploadURL(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	var body struct {
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.GetResumeUploadCredential(ctx, &recruitmentv1.GetResumeUploadCredentialRequest{
		FileName:    body.FileName,
		ContentType: body.ContentType,
	})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ResumeConfirm(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	var body recruitmentv1.ConfirmResumeUploadedRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	_, err := h.Client.ConfirmResumeUploaded(ctx, &body)
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ResumeParse(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	req := &recruitmentv1.ParseResumeToProfileRequest{}
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		fh, err := c.FormFile("resume")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resume file required"})
			return
		}
		const maxResumeBytes = 8 << 20
		if fh.Size > maxResumeBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resume file too large"})
			return
		}
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open resume"})
			return
		}
		defer f.Close()
		data, err := io.ReadAll(io.LimitReader(f, maxResumeBytes+1))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read resume"})
			return
		}
		if len(data) > maxResumeBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resume file too large"})
			return
		}
		req.FileContent = data
		req.FileName = fh.Filename
		req.ContentType = fh.Header.Get("Content-Type")
	}
	resp, err := h.Client.ParseResumeToProfile(ctx, req)
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ResumeDownloadURL(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	candID, _ := strconv.ParseInt(c.Query("candidate_user_id"), 10, 64)
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.GetResumeDownloadURL(ctx, &recruitmentv1.GetResumeDownloadURLRequest{CandidateUserId: candID})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ApplyJob(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	jobID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	_, err := h.Client.ApplyJob(ctx, &recruitmentv1.ApplyJobRequest{JobId: jobID})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ListApplications(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	jobID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.ListApplicationsForJob(ctx, &recruitmentv1.ListApplicationsForJobRequest{
		JobId: jobID, Page: int32(page), PageSize: int32(ps),
	})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Chat(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.ChatMessage(ctx, &recruitmentv1.ChatMessageRequest{Content: body.Content})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ChatHistory(c *gin.Context) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	ctx := grpcMetaCtx(c, uid.(int64), role.(string))
	resp, err := h.Client.ListChatHistory(ctx, &recruitmentv1.ListChatHistoryRequest{Page: int32(page), PageSize: int32(ps)})
	if err != nil {
		h.grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func RegisterRoutes(r *gin.Engine, h *Handler) {
	r.POST("/api/register", h.Register)
	r.POST("/api/login", h.Login)
	r.GET("/api/jobs", h.ListPublicJobs)

	auth := r.Group("/api")
	auth.Use(h.authMiddleware())
	auth.GET("/hr/jobs", h.ListMyJobs)
	auth.POST("/hr/jobs", h.CreateJob)
	auth.PATCH("/hr/jobs/:id", h.UpdateJob)
	auth.POST("/hr/jobs/:id/archive", h.ArchiveJob)
	auth.GET("/hr/jobs/:id/applications", h.ListApplications)
	auth.GET("/hr/resume_download", h.ResumeDownloadURL)
	auth.POST("/hr/chat", h.Chat)
	auth.GET("/hr/chat/history", h.ChatHistory)

	auth.GET("/me/profile", h.GetProfile)
	auth.PUT("/me/profile", h.UpdateProfile)
	auth.POST("/me/resume/upload_url", h.ResumeUploadURL)
	auth.POST("/me/resume/confirm", h.ResumeConfirm)
	auth.POST("/me/resume/parse", h.ResumeParse)
	auth.POST("/jobs/:id/apply", h.ApplyJob)
}

func DialLogic(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
