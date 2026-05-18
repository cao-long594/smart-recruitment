import fs from "fs";

function parseExam(path) {
  const text = fs.readFileSync(path, "utf8");
  const re =
    /paragraph \[ref=e\d+\]: "(\d+)\."\s*\n\s*- paragraph \[ref=e\d+\]:\s*\n\s*- generic \[ref=e\d+\]: 【([^】]+)】\s*\n\s*- generic \[ref=e\d+\]: ([^\n]+)（\d+分）/g;
  const out = [];
  let m;
  while ((m = re.exec(text)) !== null) {
    out.push({ n: +m[1], type: m[2], q: m[3].trim() });
  }
  return out;
}

const e6 = parseExam("exam6-snapshot.md");
const e7 = parseExam("exam7-snapshot.md");
fs.writeFileSync(
  "exam-questions.json",
  JSON.stringify({ exam6: e6, exam7: e7 }, null, 2),
  "utf8"
);
console.log(
  JSON.stringify(
    { exam6: e6.length, exam7: e7.length, written: "exam-questions.json" },
    null,
    2
  )
);
