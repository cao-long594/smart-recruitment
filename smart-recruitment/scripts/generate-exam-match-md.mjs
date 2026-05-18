import fs from "fs";

const { exam6, exam7 } = JSON.parse(fs.readFileSync("exam-questions.json", "utf8"));

/** 课程「二、三」节中的链接顺序（与训练营课件一致） */
const RESOURCES = [
  {
    id: "async-promise",
    title: "异步编程和 Promise 参考教程",
    url: "https://campus.wps.cn/contentpreview/2bd56646-83c4-432a-8f91-4f082ffed04e",
  },
  {
    id: "js-object",
    title: "JS 对象、原型和类",
    url: "https://campus.wps.cn/contentpreview/6283bb46-d610-49fb-aaed-391a057c8fce",
  },
  {
    id: "design-patterns",
    title: "设计模式（Human Patterns 中文版）",
    url: "https://lolipopj.github.io/design-patterns-for-humans-zh/",
  },
  {
    id: "xhr-fetch",
    title: "从 XMLHttpRequest 到 Fetch、SSE、WebSocket",
    url: "https://campus.wps.cn/contentpreview/cc1e203f-ebb8-42f9-acaa-e05eeaf7cf42",
  },
  {
    id: "nodejs",
    title: "现代 Node.js 参考教程",
    url: "https://campus.wps.cn/contentpreview/7b0e2116-2fbe-4f5b-9202-4db92696345c",
  },
  {
    id: "webpack",
    title: "DAY04 - Webpack 基础入门",
    url: "https://campus.wps.cn/contentpreview/a70c896c-3c2e-47ca-87e6-382a5ceadc42",
  },
  {
    id: "vite",
    title: "Vite 介绍（金山文档）",
    url: "https://365.kdocs.cn/l/coN1ilZbBh7I?from=koa",
  },
  {
    id: "ts-ruanyifeng",
    title: "阮一峰 TypeScript 教程",
    url: "https://wangdoc.com/typescript/",
  },
  {
    id: "typescript",
    title: "TypeScript 参考教程",
    url: "https://campus.wps.cn/contentpreview/a5baaad3-5c34-4521-8847-881ea9e5cd10",
  },
  {
    id: "http",
    title: "http&https协议（HTTP 原理与实践）",
    url: "https://campus.wps.cn/contentpreview/898a5f0c-9441-4243-b9f1-db7e65091280",
  },
  {
    id: "whistle",
    title: "Whistle（金山文档）",
    url: "https://365.kdocs.cn/l/cv1kSuuEihXc",
  },
  {
    id: "api-doc",
    title: "API 文档与接口测试（课件列表说明，无独立外链）",
    url: "https://campus.wps.cn/contentpreview/7e5a0755-818b-46a6-ba23-8071e3a634a6#三、-HTTP入门和React开发",
  },
  {
    id: "react",
    title: "现代 React 参考教程",
    url: "https://campus.wps.cn/contentpreview/b37c01f8-331b-49fa-82fa-64617e3308a6",
  },
  {
    id: "react-router",
    title: "React Router v7 参考教程",
    url: "https://campus.wps.cn/contentpreview/b6619bc4-0aa7-4e35-b096-094d1078b779",
  },
  {
    id: "zustand",
    title: "Zustand 状态管理参考教程",
    url: "https://campus.wps.cn/contentpreview/c87f1a45-d3f9-4a14-8e41-d77b1f4fe9fc",
  },
  {
    id: "out-of-scope",
    title: "题目超出「二、JavaScript 进阶」「三、HTTP 与 React」课件链时的说明",
    url: "https://campus.wps.cn/contentpreview/7e5a0755-818b-46a6-ba23-8071e3a634a6",
  },
];

const rid = (id) => RESOURCES.findIndex((r) => r.id === id);

/**
 * @returns {{ id: string, chapter: string, section: string, reason: string, conf: number, sort: string }}
 */
function classify(q, exam, num) {
  const t = `${q}`;

  if (/大语言模型|LLM|API Key|敏感信息.*泄漏/i.test(t)) {
    return {
      id: "out-of-scope",
      chapter: "（通识/安全）",
      section: "非前端「二、三」节课件内容",
      reason: "题干为 AI/密钥安全通识，与所列前端教程链无直接章节对应",
      conf: 35,
      sort: "05",
    };
  }
  if (/Koa|洋葱|中间件.*Koa/i.test(t)) {
    return {
      id: "nodejs",
      chapter: "第十五章 Web 开发 — Koa 框架",
      section: "15.3 中间件机制（洋葱模型）",
      reason: "题目考查 Koa 中间件执行顺序",
      conf: 90,
      sort: "12",
    };
  }
  if (/编程式导航|useNavigate/.test(t)) {
    return {
      id: "react-router",
      chapter: "6. 导航",
      section: "6.3 编程式导航——useNavigate",
      reason: "题目考查路由编程式跳转 API",
      conf: 91,
      sort: "28",
    };
  }
  if (/toRef|toRefs|组合式|Vue 3|Vue3/.test(t) || (/Vue/.test(t) && /指令|响应式|provide|inject|生命周期/.test(t))) {
    return {
      id: "typescript",
      chapter: "第七章 在 Vue3 中的实践",
      section: "7.2 学习相关语法",
      reason: "Vue 相关题在「三」中无单独 Vue 教程，对应 TS 课件中的 Vue3 章节",
      conf: 80,
      sort: "50",
    };
  }
  if (/自定义\s*Hook|自定义Hook|关于自定义 Hook/i.test(t)) {
    return {
      id: "react",
      chapter: "（Hooks 进阶）",
      section: "自定义 Hook 约定与副作用",
      reason: "题目考查 React 自定义 Hook 规范",
      conf: 84,
      sort: "42",
    };
  }
  if (
    /CSS|伪类|选择器|text-overflow|nth-child|innerHTML|innerText|querySelector|getElementsBy|事件委托|localStorage|sessionStorage|BOM|地址栏|元素的CSS|类操作|自定义属性/.test(
      t
    )
  ) {
    return {
      id: "out-of-scope",
      chapter: "「一、前端开发基础」等",
      section: "HTML/CSS/DOM/Web APIs（不在本次「二、三」链接清单内）",
      reason: "题干为 DOM/BOM/存储/CSS 等，对应训练营第一节资料而非本次所列二、三链接",
      conf: 38,
      sort: "06",
    };
  }
  if (/定时器|setInterval|setTimeout/.test(t) && !/Promise|输出|执行结果/.test(t)) {
    return {
      id: "async-promise",
      chapter: "第一章 JavaScript 为什么需要异步",
      section: "1.3 常见的异步场景",
      reason: "题目考查定时器与异步基础（非输出顺序题）",
      conf: 68,
      sort: "15",
    };
  }

  if (/FormData/.test(t)) {
    return {
      id: "xhr-fetch",
      chapter: "第三章 Fetch API",
      section: "3.6 POST 请求 / Request 与 body",
      reason: "FormData 常用于 fetch/xhr 提交表单数据",
      conf: 80,
      sort: "91",
    };
  }
  if (/文件上传|enctype|multipart/.test(t)) {
    return {
      id: "http",
      chapter: "2. HTTP消息结构",
      section: "请求体与表单编码（multipart 等）",
      reason: "表单编码与 HTTP 报文相关",
      conf: 72,
      sort: "22",
    };
  }

  if (/Whistle/i.test(t)) {
    return {
      id: "whistle",
      chapter: "（金山文档）",
      section: "代理与调试命令",
      reason: "题目直接考查 Whistle 工具",
      conf: 92,
      sort: "10",
    };
  }
  if (/Git|提交规范|feat|perf|fix|commit|conventional/i.test(t)) {
    return {
      id: "out-of-scope",
      chapter: "「一、前端开发基础」Git 资料",
      section: "不在本次「二、三」课件链接清单内",
      reason: "Git 提交规范对应训练营第一节链接，非第二节起的资源表",
      conf: 45,
      sort: "98",
    };
  }
  if (/HTTP\s*状态码|状态码\s*404|状态码\s*400|Content-Type|HTTPS|HTTP\s*协议|请求头/i.test(t)) {
    return {
      id: "http",
      chapter: "2. HTTP消息结构 / 4. 状态码与缓存",
      section: "与题目中的报文、状态码或头部相关",
      reason: "考查 HTTP 报文、状态码或头部语义",
      conf: 88,
      sort: "20",
    };
  }
  if (/React Router|Router|Link 和 NavLink|NavLink|按需加载.*路由|懒加载.*路由|createBrowserRouter|路由.*根组件|BrowserRouter/i.test(t)) {
    return {
      id: "react-router",
      chapter: "3. 创建路由——createBrowserRouter / 6. 导航",
      section: "路由配置、Link/NavLink、懒加载",
      reason: "题目围绕 React Router 使用或原理",
      conf: 93,
      sort: "30",
    };
  }
  if (
    /React|useState|useEffect|useMemo|useCallback|useRef|useContext|Context|Fragment|memo|Fiber|高阶组件|Hooks|hooks|Redux|reducer|action|props|state|组件|渲染|key\s*的|单项数据流/i.test(
      t
    )
  ) {
    let section = "第五章 State / 第八章 useEffect / 性能优化相关节";
    if (/useEffect|副作用|依赖项|清理函数/.test(t)) section = "第八章 副作用与 useEffect";
    else if (/useState|state|setState|不可变/.test(t)) section = "第五章 State 与事件处理";
    else if (/useMemo|useCallback|memo|优化/.test(t)) section = "性能优化与 useMemo/useCallback 相关章节";
    else if (/Context|useContext/.test(t)) section = "Context 与全局状态相关章节";
    else if (/Redux|reducer|action/.test(t)) section = "与状态管理概念衔接（可与 Zustand 对照）";
    else if (/key/.test(t)) section = "第六章 列表与 Key";
    return {
      id: "react",
      chapter: "现代 React 参考教程（多章综合）",
      section,
      reason: "题目考查 React 组件、Hooks 或数据流",
      conf: 86,
      sort: "40",
    };
  }
  if (/Vue/.test(t)) {
    return {
      id: "typescript",
      chapter: "第七章 在 Vue3 中的实践",
      section: "7.2 学习相关语法 / 组合式 API",
      reason: "课件「三」未单列 Vue 教程；TypeScript 参考教程含 Vue3 实践章节",
      conf: 78,
      sort: "50",
    };
  }
  if (/Zustand|zustand/.test(t)) {
    return {
      id: "zustand",
      chapter: "3. 快速上手 — 5. 选择器",
      section: "与轻量全局状态相关小节",
      reason: "题目点名 Zustand 或与其模式一致",
      conf: 90,
      sort: "55",
    };
  }
  if (/全局状态管理库|Redux\)/.test(t)) {
    return {
      id: "zustand",
      chapter: "1. 为什么选择 Zustand",
      section: "与 Redux 对比的动机",
      reason: "题目讨论全局状态库选型，Zustand 教程开篇对比 Redux",
      conf: 72,
      sort: "56",
    };
  }
  if (/Webpack|webpack|babel-loader|css-loader|devtool|externals|optimization|resolve|loader/.test(t)) {
    return {
      id: "webpack",
      chapter: "DAY04 课件目录",
      section: "与 webpack 配置项对应的小节标题",
      reason: "题目考查 Webpack 配置或打包概念",
      conf: 90,
      sort: "60",
    };
  }
  if (/Vite|vite/.test(t)) {
    return {
      id: "vite",
      chapter: "（金山文档）",
      section: "Vite 入门与配置",
      reason: "题目考查 Vite",
      conf: 85,
      sort: "65",
    };
  }
  if (/TypeScript|泛型|接口类型|类型断言/.test(t)) {
    return {
      id: "typescript",
      chapter: "4. TypeScript 基础 / 5. 泛型",
      section: "类型系统相关小节",
      reason: "题目考查 TS 类型系统",
      conf: 88,
      sort: "70",
    };
  }
  if (/Node\.js|nodejs|内置模块|npm|npx|fs\.|path\.|http\.|模块|CommonJS|ESM|cluster|Buffer|stream/i.test(t)) {
    let chapter = "第一—十二章 内置模块与 npm";
    if (/内置模块|path|url|http|dayjs/.test(t)) chapter = "第五章 模块系统 / 第十二章 其他常用模块 / 第十章 http";
    if (/HTTP 请求|发起请求/.test(t) && /Node/i.test(t))
      chapter = "第十章 内置模块 — http 网络 — 10.5 发起 HTTP 请求";
    return {
      id: "nodejs",
      chapter,
      section: "与 Node 运行时或内置包相关",
      reason: "题目考查 Node 能力或内置模块辨析",
      conf: 87,
      sort: "80",
    };
  }
  if (/axios|fetch|xhr|XMLHttpRequest|AJAX|WebSocket|SSE|CORS|跨域|AbortController|代理请求/i.test(t)) {
    return {
      id: "xhr-fetch",
      chapter: "第二—六章 XHR / Fetch / SSE / WebSocket",
      section: "与网络 API 或跨域相关小节",
      reason: "题目考查浏览器网络 API、跨域或请求控制",
      conf: 88,
      sort: "90",
    };
  }
  if (/apply和call|call和apply/.test(t)) {
    return {
      id: "js-object",
      chapter: "第三章 对象方法与 this",
      section: "3.2 this 的核心规则 / 函数调用方式",
      reason: "题目考查 call/apply 与 this",
      conf: 88,
      sort: "112",
    };
  }
  if (/typeof|类型转换|隐式转换/.test(t) && /输出|结果/.test(t)) {
    return {
      id: "js-object",
      chapter: "第一章 对象基础 / 第二章 引用与复制",
      section: "原始值与运算（弱关联章节）",
      reason: "输出题涉及类型运算，更接近 JS 语言基础而非 Promise 专章",
      conf: 58,
      sort: "113",
    };
  }
  if (/点击按钮|setList|列表.*更新|useState.*\[|map\(item/.test(t)) {
    return {
      id: "react",
      chapter: "第七章 State 管理进阶",
      section: "7.2 / 7.3 不可变更新数组与对象",
      reason: "题干描述典型 React 列表与 state 不可变更新问题",
      conf: 82,
      sort: "45",
    };
  }
  if (/下面.*输出|执行结果是什么|代码的执行结果|输出顺序|输出的结果是/.test(t)) {
    return {
      id: "async-promise",
      chapter: "第八章 微任务与事件循环",
      section: "8.5 经典面试题：输出顺序",
      reason: "题干为输出结果类题目，与异步执行顺序强相关",
      conf: 74,
      sort: "101",
    };
  }
  if (/console\.log/.test(t) && !/typeof/.test(t)) {
    return {
      id: "async-promise",
      chapter: "第八章 微任务与事件循环",
      section: "8.5 经典面试题：输出顺序",
      reason: "题干含 console.log 输出类考查，多为异步/顺序相关",
      conf: 70,
      sort: "102",
    };
  }
  if (/Promise|async|await|setTimeout|微任务|宏任务|事件循环|回调地狱|then\(|catch/.test(t)) {
    return {
      id: "async-promise",
      chapter: "第三—九章 Promise 与 async/await / 第八章 微任务",
      section: "异步语法与执行顺序相关",
      reason: "题目考查 Promise、async/await 或事件循环输出",
      conf: 90,
      sort: "100",
    };
  }
  if (/原型|constructor|class\s|this\.|instanceof|对象数据类型|深拷贝|浅拷贝/.test(t)) {
    return {
      id: "js-object",
      chapter: "第一—十一章 对象与原型链",
      section: "对象类型、方法、this 或原型相关小节",
      reason: "题目考查对象模型、this 或原型",
      conf: 86,
      sort: "110",
    };
  }
  if (/设计模式|单例|工厂模式|观察者|适配器|策略模式|装饰器模式/.test(t)) {
    return {
      id: "design-patterns",
      chapter: "（站点总览）",
      section: "对应创建型/结构型/行为型章节",
      reason: "题目考查经典设计模式",
      conf: 82,
      sort: "120",
    };
  }
  if (/preventDefault|stopPropagation|DOMContentLoaded|加载完成了 HTML|外部资源|默认事件|取消默认|阻止事件冒泡|阻止冒泡/.test(t)) {
    return {
      id: "out-of-scope",
      chapter: "浏览器事件与页面生命周期",
      section: "见「一、」Web APIs / 调试资料",
      reason: "页面 load/DOM 事件属 Web 基础，不在本次「二、三」课件链接内",
      conf: 48,
      sort: "130",
    };
  }
  if (/YAPI|Apifox|APIFox|接口测试|Vite\s*proxy|mock/.test(t)) {
    return {
      id: "api-doc",
      chapter: "课程列表文字说明",
      section: "API 文档（YAPI，APIfox）与接口测试（Vite proxy）",
      reason: "与课件第三节文字说明一致（无单独教程页）",
      conf: 70,
      sort: "140",
    };
  }
  if (/阮一峰|wangdoc.*ts/i.test(t)) {
    return {
      id: "ts-ruanyifeng",
      chapter: "（在线教程目录）",
      section: "与题目主题对应的章节",
      reason: "系统 TS 语法可参考阮一峰教程目录",
      conf: 80,
      sort: "150",
    };
  }

  return {
    id: "out-of-scope",
    chapter: "（待人工）",
    section: "与「二、三」各链接标题无明显关键词重合",
    reason: "启发式规则未命中；请对照卷面代码或选项人工归入最接近章节",
    conf: 32,
    sort: "999",
  };
}

const rows = [];

for (const { n, type, q } of exam6) {
  const m = classify(q, 6, n);
  rows.push({
    exam: "第六次",
    label: `6-Q${n}`,
    type,
    q,
    ...m,
    resourceIdx: rid(m.id),
  });
}
for (const { n, type, q } of exam7) {
  const m = classify(q, 7, n);
  rows.push({
    exam: "第七次",
    label: `7-Q${n}${type === "多选题" ? "(多选)" : ""}`,
    type,
    q,
    ...m,
    resourceIdx: rid(m.id),
  });
}

rows.sort((a, b) => {
  if (a.resourceIdx !== b.resourceIdx) return a.resourceIdx - b.resourceIdx;
  if (a.chapter !== b.chapter) return a.chapter.localeCompare(b.chapter, "zh-CN");
  if (a.section !== b.section) return a.section.localeCompare(b.section, "zh-CN");
  return a.label.localeCompare(b.label, "zh-CN");
});

let md = `# 第六、七次前端摸底 — 题目与「二、三」节教程章节映射

> 数据来源：通过 Playwright 连接浏览器读取 [金山考试个人中心](https://campus.wps.cn/usercenter/home) 中 **武科大 - 第六次前端摸底测试**、**第七次前端摸底测试** 的「查看详情」页面；教程目录来自训练营课件内各 **金山课件 / 外链** 的侧边栏快照（2026-05-12）。  
> 题号前缀：**6-Q** 表示第六次第 n 题，**7-Q** 表示第七次第 n 题（第七次含一道多选题标注「多选」）。  
> **置信度** 为基于题干关键词与教程章节语义的启发式匹配，非模型概率；「泛输出题」等建议结合卷面代码人工复核。

---

`;

for (const res of RESOURCES) {
  const sub = rows.filter((r) => r.id === res.id);
  if (sub.length === 0) continue;
  md += `## ${res.title}\n\n链接：${res.url}\n\n`;
  const byChapter = new Map();
  for (const r of sub) {
    const k = `${r.chapter}|||${r.section}`;
    if (!byChapter.has(k)) byChapter.set(k, []);
    byChapter.get(k).push(r);
  }
  for (const [k, list] of byChapter) {
    const [chapter, section] = k.split("|||");
    md += `### ${chapter}\n\n#### ${section}\n\n`;
    md += `| 题号 | 题目 | 匹配原因 | 置信度 |\n| --- | --- | --- | --- |\n`;
    for (const r of list) {
      const qq = r.q.replace(/\|/g, "\\|");
      md += `| ${r.label} | ${qq} | ${r.reason} | ${r.conf}% |\n`;
    }
    md += "\n";
  }
  md += "---\n\n";
}

md += `## 统计摘要\n\n`;
const counts = {};
for (const r of rows) {
  counts[r.id] = (counts[r.id] || 0) + 1;
}
for (const res of RESOURCES) {
  const c = counts[res.id] || 0;
  if (c) md += `- **${res.title}**：${c} 题\n`;
}

fs.writeFileSync("考试题目与教程章节映射-第六七次.md", md, "utf8");
console.log("Wrote 考试题目与教程章节映射-第六七次.md", "rows", rows.length);
