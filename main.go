package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v2"
)

var mcpServer *server.MCPServer
var toolRegistry = map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error){}

func main() {
	s := server.NewMCPServer("mcp-redteam", "1.0.0")
	mcpServer = s

	s.AddTool(mcp.NewTool("subfinder",
		mcp.WithDescription("Subfinder is a subdomain discovery tool that discovers subdomains for websites by using passive online sources"),
		mcp.WithString("domain", mcp.Required())), subfinderHandler)
	toolRegistry["subfinder"] = subfinderHandler

	s.AddTool(mcp.NewTool("amass",
		mcp.WithDescription("In-depth Attack Surface Mapping and Asset Discovery"),
		mcp.WithString("domain", mcp.Required())), amassHandler)
	toolRegistry["amass"] = amassHandler

	s.AddTool(mcp.NewTool("httpx",
		mcp.WithDescription("httpx is a fast and multi-purpose HTTP toolkit that allows running multiple probes"),
		mcp.WithString("hosts")), httpxHandler)
	toolRegistry["httpx"] = httpxHandler

	s.AddTool(mcp.NewTool("dnsx",
		mcp.WithDescription("dnsx is a fast and multi-purpose DNS toolkit allow to run multiple probes"),
		mcp.WithString("domains")), dnsxHandler)
	toolRegistry["dnsx"] = dnsxHandler

	s.AddTool(mcp.NewTool("katana", mcp.WithDescription("Katana is a fast crawler focused on execution in automation"), mcp.WithString("urls")), katanaHandler)
	toolRegistry["katana"] = katanaHandler

	s.AddTool(mcp.NewTool("nuclei", mcp.WithDescription("Nuclei is a fast, template based vulnerability scanner"), mcp.WithString("urls")), nucleiHandler)
	toolRegistry["nuclei"] = nucleiHandler

	s.AddTool(mcp.NewTool("semgrep", mcp.WithDescription("Semgrep is a fast, open-source, static analysis tool"), mcp.WithString("repo", mcp.Required())), semgrepHandler)
	toolRegistry["semgrep"] = semgrepHandler

	s.AddTool(mcp.NewTool("shodan", mcp.WithDescription("Search Engine for the Internet of Everything"), mcp.WithString("query", mcp.Required())), shodanHandler)
	toolRegistry["shodan"] = shodanHandler

	s.AddTool(mcp.NewTool("fofa", mcp.WithDescription("A search engine that maps global cyberspace"), mcp.WithString("query", mcp.Required())), fofaHandler)
	toolRegistry["fofa"] = fofaHandler

	s.AddTool(mcp.NewTool("report", mcp.WithDescription("A markdown report module"), mcp.WithString("input", mcp.Required())), reportHandler)
	toolRegistry["report"] = reportHandler

	s.AddTool(mcp.NewTool("pipeline",
		mcp.WithString("yaml_template"),
		mcp.WithString("params", mcp.Required(), mcp.Description("YAML payload containing name and tasks")),
	), pipelineHandler)
	toolRegistry["pipeline"] = pipelineHandler

	s.AddTool(mcp.NewTool("expand-prompt",
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Prompt like 'scan domain.net using amass and httpx'")),
	), expandPromptHandler)
	toolRegistry["expand-prompt"] = expandPromptHandler

	s.AddTool(mcp.NewTool("plan-and-execute",
		mcp.WithString("prompt", mcp.Required(), mcp.Description("Prompt to generate and execute a pipeline")),
	), planAndExecuteHandler)
	toolRegistry["plan-and-execute"] = planAndExecuteHandler

	server.ServeStdio(s)
}

func runCommand(name string, args []string, input []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // 60 secs timeout
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return "", err
	}
	for _, line := range input {
		stdin.Write([]byte(line + "\n"))
	}
	stdin.Close()
	out, err := io.ReadAll(stdout)
	if err != nil {
		return "", err
	}
	cmd.Wait()
	return string(out), nil
}

func subfinderHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain := req.Params.Arguments["domain"].(string)
	out, err := exec.Command("subfinder", "-d", domain, "-silent").Output()
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(out)), nil
}

func amassHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain := req.Params.Arguments["domain"].(string)
	out, err := exec.Command("amass", "enum", "-timeout", "1", "-d", domain).Output()
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(out)), nil
}

func httpxHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hosts := strings.Split(req.Params.Arguments["hosts"].(string), "\n")
	out, err := runCommand("httpx", []string{"-silent"}, hosts)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(out), nil
}

func dnsxHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domains := strings.Split(req.Params.Arguments["domains"].(string), "\n")
	out, err := runCommand("dnsx", nil, domains)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(out), nil
}

func katanaHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	urls := strings.Split(req.Params.Arguments["urls"].(string), "\n")
	out, err := runCommand("katana", []string{"-silent"}, urls)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(out), nil
}

func nucleiHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	urls := strings.Split(req.Params.Arguments["urls"].(string), "\n")
	// TODO: strip urls from https so nuclei can look up other ports
	out, err := runCommand("nuclei", []string{"-silent", "-as", "-me"}, urls)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(out), nil
}

func semgrepHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := req.Params.Arguments["repo"].(string)
	out, err := exec.Command("semgrep", "--config=p/owasp-top-ten", repo).Output()
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(out)), nil
}

func shodanHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.Params.Arguments["query"].(string)
	out, err := exec.Command("shodan", "search", query).Output()
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(out)), nil
}

func fofaHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.Params.Arguments["query"].(string)
	out, err := exec.Command("fofa", "query", query).Output()
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(out)), nil
}

func reportHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content := req.Params.Arguments["input"].(string)
	lines := strings.Split(content, "\n")
	var sb strings.Builder
	sb.WriteString("# MCP Security Report\n\n")
	for _, line := range lines {
		sb.WriteString("- " + line + "\n")
	}
	return mcp.NewToolResultText(sb.String()), nil
}

func expandPromptHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt := strings.ToLower(req.Params.Arguments["prompt"].(string))
	domain := "example.com"
	for _, word := range strings.Fields(prompt) {
		if strings.Contains(word, ".") {
			domain = word
			break
		}
	}
	yaml := "name: generated\ndescription: pipeline generated from prompt\n\ntasks:\n"
	if strings.Contains(prompt, "amass") {
		yaml += fmt.Sprintf("  - name: amass\n    tool: amass\n    args:\n      domain: %s\n", domain)
	}
	if strings.Contains(prompt, "subfinder") {
		yaml += fmt.Sprintf("  - name: subfinder\n    tool: subfinder\n    args:\n      domain: %s\n", domain)
	}
	if strings.Contains(prompt, "httpx") {
		yaml += fmt.Sprintf("  - name: httpx\n    tool: httpx\n    args:\n      hosts: %s\n", domain)
	}
	if strings.Contains(prompt, "nuclei") {
		yaml += fmt.Sprintf("  - name: nuclei\n    tool: nuclei\n    args:\n      urls: https://%s\n", domain)
	}
	if strings.Contains(prompt, "dnsx") {
		yaml += fmt.Sprintf("  - name: dnsx\n    tool: dnsx\n    args:\n      domains: %s\n", domain)
	}
	if yaml == "name: generated\ndescription: pipeline generated from prompt\n\ntasks:\n" {
		yaml += fmt.Sprintf("  - name: default\n    tool: httpx\n    args:\n      hosts: %s\n", domain)
	}
	return mcp.NewToolResultText(yaml), nil
}

func planAndExecuteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt := req.Params.Arguments["prompt"].(string)

	expandReq := mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
	}
	expandReq.Params.Name = "expand-prompt"
	expandReq.Params.Arguments = map[string]interface{}{"prompt": prompt}

	yamlResult, err := expandPromptHandler(ctx, expandReq)
	if err != nil {
		return nil, fmt.Errorf("failed to expand prompt: %w", err)
	}
	yaml := yamlResult.Content[0].(mcp.TextContent)

	pipeReq := mcp.CallToolRequest{
		Request: mcp.Request{Method: "tools/call"},
	}
	pipeReq.Params.Name = "pipeline"
	pipeReq.Params.Arguments = map[string]interface{}{"yaml_template": "generated", "params": yaml}

	return pipelineHandler(ctx, pipeReq)
}

func pipelineHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rawYAML, ok := req.Params.Arguments["params"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid YAML string passed in 'params'")
	}
	var parsed struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Tasks       []struct {
			Name string                 `yaml:"name"`
			Tool string                 `yaml:"tool"`
			Args map[string]interface{} `yaml:"args"`
		} `yaml:"tasks"`
	}
	if err := yaml.Unmarshal([]byte(rawYAML), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	var resultBuilder strings.Builder
	for _, task := range parsed.Tasks {
		handler, exists := toolRegistry[task.Tool]
		if !exists {
			resultBuilder.WriteString(fmt.Sprintf("[!] Unknown tool: %s\n", task.Tool))
			continue
		}
		toolReq := mcp.CallToolRequest{
			Request: mcp.Request{Method: "tools/call"},
		}
		toolReq.Params.Name = task.Tool
		toolReq.Params.Arguments = task.Args
		result, err := handler(ctx, toolReq)
		if err != nil {
			resultBuilder.WriteString(fmt.Sprintf("[!] Error running %s: %v\n", task.Tool, err))
			continue
		}
		if len(result.Content) > 0 {
			text := result.Content[0].(*mcp.TextContent)
			resultBuilder.WriteString(fmt.Sprintf("## %s\n%v\n\n", task.Name, text))
		}
	}
	return mcp.NewToolResultText(resultBuilder.String()), nil
}
