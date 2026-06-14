# zenodo-cli 领域术语表

## 核心守则

**zenodo-cli 是 Zenodo 网页版的替代品**，额外提供自动化便利和 agent 友好。

## 核心概念

**zenodo-cli** — Zenodo/InvenioRDM 的本地 CLI 工具，用于管理沉积记录、文件上传下载和完整 InvenioRDM API 访问。安装后的二进制文件是 `zenodo`。

## 用户

**研究者** — 需要管理 Zenodo 沉积记录、上传数据集的人。

**Agent** — 自动化代理（CI、脚本、数据管道）。需要确定性行为。

## 命令设计决策

**安全门控** — `--read-only`、`--dry-run`、`--confirm` 三级安全机制。

**原始 API 逃生舱** — `zenodo api` 允许直接调用任何 InvenioRDM 端点。
