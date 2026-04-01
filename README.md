<p align="center">
	<img src="assets/jotform-logo.png" alt="Jotform CLI logo" width="760" />
</p>

# Jotform CLI: The AI-Native Data Collection Layer

## 1. Executive Summary
The software landscape is shifting from Graphical User Interfaces (GUI) to **Agentic Workflows**. Developers and AI Agents (such as **Claude Code**, **OpenAI Operator**, and **GitHub Copilot**) require a high-velocity, terminal-based interface to create, manage, and deploy data collection points.

The **Jotform CLI** is designed to transform Jotform from a "No-Code Form Builder" into a core **Developer Infrastructure Component**, enabling seamless integration into CI/CD pipelines and autonomous AI environments via the **Model Context Protocol (MCP)**.

---

## 2. Core Pillars

### A. AI-Agent Compatibility (The "Agentic" Layer)
* **MCP Server Integration:** A built-in Model Context Protocol server that allows **Claude Desktop** and other LLMs to "see" Jotform as a native tool in their environment.
* **Prompt-to-Form:** A command-driven engine where an agent can execute `jotform ai generate "Survey for Sür app users"` and receive a validated Jotform JSON schema.
* **Headless Deployment:** Enabling agents to deploy and update forms instantly using `jotform deploy --schema`.

### B. Developer Experience (DX) & CI/CD
* **Form-as-Code (FaC):** Export and import form definitions as **YAML/JSON**. This allows developers to version-control their forms in Git, enabling "Rollbacks" and "Pull Request" reviews for form logic changes.
* **Single Binary (Go-based):** Built with Golang for zero-dependency installation across macOS, Linux, and Windows.
* **Submission Stream:** Pipe form submissions directly into other CLI tools or local databases (e.g., `jotform submissions --watch | jq .`).

### C. Open Source Strategy
* **Community-Driven:** The CLI core will be Open Source to build trust within the developer community and ensure security transparency for API key management.
* **Extensibility:** A plugin architecture that allows the community to build custom "Formatters" (e.g., exporting Jotform data directly into Supabase, Firebase, or PostgreSQL).

---

## 3. Technical Architecture & Command Mapping

### Module 1: `jotform auth`
* `login / logout`: Securely manage API Keys using system keychains.
* `whoami`: Verify account status and API usage limits.

### Module 2: `jotform forms` (The Infrastructure Layer)
* `list`: List all active forms with metadata.
* `get [id] --format json`: Fetch form structure for local editing.
* `create --file [path]`: Deploy a new form from a local definition file.
* `sync`: Pull remote form changes to local version-controlled files.

### Module 3: `jotform ai` (The AI Bridge)
* `generate-schema "[prompt]"`: Uses LLM reasoning to output a valid Jotform-compatible JSON structure.
* `analyze [id]`: Feeds form structure to an agent to suggest UX improvements or logic optimizations.

### Module 4: `jotform mcp`
* `start-server`: Launches the MCP server, exposing `create_form`, `list_forms`, and `get_submissions` as tools for **Claude Code**.

---

## 4. Strategic Impact for Jotform

1.  **First-Mover Advantage:** By releasing an MCP-compliant CLI, Jotform becomes the *de facto* data collection tool for the millions of developers moving to AI-assisted coding.
2.  **Enterprise Adoption:** Version-controlling forms (Form-as-Code) solves a major pain point for engineering teams (like those at **İmarAnaliz**) who require audit trails for form changes.
3.  **Ecosystem Expansion:** Transitioning from a SaaS product to a "Developer Utility" increases "stickiness" and reduces churn among technical users.

---

## 5. Roadmap
* **Phase 1:** Core Go-based CLI (Auth, CRUD, JSON Export/Import).
* **Phase 2:** AI-Link for `generate-schema` functionality.
* **Phase 3:** **Model Context Protocol (MCP)** implementation for Claude/Agent integration.
* **Phase 4:** Official Open Source release and developer outreach.