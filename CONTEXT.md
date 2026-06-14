# zenodo-cli 领域术语表

## 核心守则

**zenodo-cli 是 Zenodo 网页版的替代品**，额外提供自动化便利和 agent 友好。能参照 Zenodo/InvenioRDM API 做法的自行决定，只询问 UX 相关的决策。

## 核心概念

**zenodo-cli** — Zenodo/InvenioRDM 的本地 CLI 工具，用于管理沉积记录、文件上传下载和完整 InvenioRDM API 访问。仓库名 `zenodo-cli`，安装后的二进制文件是 `zenodo`。

**InvenioRDM** — Zenodo 2023 年迁移后的新基础设施 API（`/api/records`）。zenodo-cli 基于此 API 构建，旧的 `/api/deposit/depositions` 端点已返回 404。

## 用户

**研究者** — 需要管理 Zenodo 沉积记录、上传数据集、发布研究成果的人。可能有多个账户（生产、沙盒），通过 profile 管理。

**Agent** — 自动化代理（CI、脚本、数据管道）。需要确定性行为：稳定 JSON 输出、可预测退出码、安全门控。典型场景：批量上传数据集、自动发布记录。

## 命令设计决策

**Profile 管理** — 用户可能有多个 Zenodo 账户（生产、沙盒），通过 `--profile` 切换。默认 profile 是用户添加的第一个账户。

**沙盒模式** — `--sandbox` 或 `ZENODO_SANDBOX=1` 切换到 sandbox.zenodo.org，不触碰生产数据。适合测试和开发。

**安全门控** — 三级安全机制：
- `--read-only` — 全局开关，阻止所有远程修改（适合脚本/Agent）
- `--dry-run` — 操作级，只预览不执行
- `--confirm` — 操作级，确认高风险操作（删除、发布不可逆）

**发布不可逆** — `records publish` 是不可逆操作，必须 `--confirm`。这是 Zenodo 平台本身的约束。

**原始 API 逃生舱** — `zenodo api` 命令允许直接调用任何 InvenioRDM 端点（GET/POST/PUT），当预置命令不覆盖某个场景时使用。路径前缀 `/api` 自动补全。

**输出格式** — 三级输出：
- 默认 — 简洁关键信息
- `--full` — 完整字段（覆盖 `--compact`）
- `--json` — 自动化/Agent 用

**事件流** — `--events` 在 stderr 输出 NDJSON 进度事件，stdout 保持 JSON 结果。适合长时间操作（大文件上传）的进度监控。
