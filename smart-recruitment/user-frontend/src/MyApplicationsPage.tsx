export function MyApplicationsPage() {
  return (
    <div className="page-shell-inner">
      <section className="page-hero">
        <div>
          <p className="eyebrow">投递记录</p>
          <h1 className="page-title">已投递岗位与状态</h1>
          <p className="page-lead">
            当前后端接口尚未开放候选人侧投递列表，本页保留入口并明确展示接入状态。
          </p>
        </div>
      </section>
      <div className="content-card empty-panel">
        <h1 className="page-title">已投递岗位与状态</h1>
        <p className="muted page-lead">
          当前后端尚未提供「候选人投递列表」接口，无法在页面中展示历史投递与状态。
        </p>
        <p className="muted">
          接入接口后，可在此展示岗位名称、投递时间与审核状态等信息。您仍可继续在岗位详情页发起投递。
        </p>
      </div>
    </div>
  )
}
