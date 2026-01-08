package handler

import (
	"fmt"
	"net/http"

	"github.com/aethersailor/subconverter-extended/internal/version"
)

func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	commitLink := ""
	if version.BuildID != "" {
		commitLink = fmt.Sprintf(`<a href="https://github.com/Aethersailor/SubConverter-Extended/commit/%s" target="_blank">%s</a>`, version.BuildID, version.BuildID)
	}
	page := fmt.Sprintf(versionPageTemplate, version.Version, commitLink, version.FormatBuildDate(version.BuildDate))
	_, _ = w.Write([]byte(page))
}

const versionPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="color-scheme" content="light dark">
    <title>SubConverter-Extended</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            /* Light Theme - 精准调优 */
            --bg-gradient: linear-gradient(135deg, #f5f7fa 0%, #e4e7eb 100%);
            --container-bg: rgba(255, 255, 255, 0.85);
            --container-border: rgba(255, 255, 255, 0.4);
            --shadow: 0 12px 40px rgba(31, 38, 135, 0.08);
            --text-primary: #1a202c;
            --text-secondary: #4a5568;
            --divider-bg: linear-gradient(90deg, transparent, rgba(0,0,0,0.08), transparent);
            --info-block-bg: rgba(0, 0, 0, 0.02);
            --info-block-border: rgba(0,0,0,0.04);
            --link-color: #3182ce;
            --link-hover: #2b6cb0;
            --header-gradient: linear-gradient(135deg, #1a202c 0%, #4a5568 100%);
        }

        @media (prefers-color-scheme: dark) {
            :root {
                /* Dark Theme - 极黑质感 */
                --bg-gradient: radial-gradient(circle at 50% 50%, #1a1b1e 0%, #000000 100%);
                --container-bg: rgba(28, 28, 30, 0.7);
                --container-border: rgba(255, 255, 255, 0.1);
                --shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
                --text-primary: #f8f9fa;
                --text-secondary: #a0aec0;
                --divider-bg: linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent);
                --info-block-bg: rgba(255, 255, 255, 0.04);
                --info-block-border: rgba(255,255,255,0.06);
                --link-color: #63b3ed;
                --link-hover: #90cdf4;
                --header-gradient: linear-gradient(135deg, #ffffff 0%, #90cdf4 100%);
            }
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: 'Outfit', system-ui, -apple-system, sans-serif;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: var(--bg-gradient);
            background-attachment: fixed;
            padding: 24px;
            color: var(--text-primary);
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
        }
        
        .container {
            background: var(--container-bg);
            backdrop-filter: blur(24px);
            -webkit-backdrop-filter: blur(24px);
            border-radius: 32px;
            padding: 40px 50px;
            max-width: 800px;
            width: 100%;
            box-shadow: var(--shadow);
            border: 1px solid var(--container-border);
            position: relative;
            animation: fadeIn 1s cubic-bezier(0.16, 1, 0.3, 1);
        }

        .container::after {
            content: "";
            position: absolute;
            inset: 0;
            border-radius: 32px;
            padding: 1px;
            background: linear-gradient(135deg, rgba(255,255,255,0.2), transparent, rgba(255,255,255,0.05));
            -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
            mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
            -webkit-mask-composite: xor;
            mask-composite: exclude;
            pointer-events: none;
        }
        
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(30px); }
            to { opacity: 1; transform: translateY(0); }
        }
        
        header {
            text-align: center;
            margin-bottom: 32px;
        }

        h1 {
            background: var(--header-gradient);
            -webkit-background-clip: text;
            background-clip: text;
            -webkit-text-fill-color: transparent;
            font-size: 3em;
            margin-bottom: 8px;
            font-weight: 700;
            letter-spacing: -0.04em;
            line-height: 1.05;
        }
        
        .subtitle {
            color: var(--text-secondary);
            font-size: 1.05em;
            font-weight: 500;
            letter-spacing: 0.1em;
            text-transform: uppercase;
            opacity: 0.6;
        }
        
        .section {
            margin: 20px 0;
            padding: 20px 25px;
            background: var(--info-block-bg);
            border-radius: 20px;
            border: 1px solid var(--info-block-border);
        }

        .section-title {
            font-size: 0.9em;
            font-weight: 700;
            color: var(--text-primary);
            margin-bottom: 15px;
            display: flex;
            align-items: center;
            gap: 8px;
            text-transform: uppercase;
            letter-spacing: 0.1em;
            opacity: 0.8;
        }

        .description {
            color: var(--text-secondary);
            font-size: 1em;
            line-height: 1.8;
            margin-bottom: 12px;
            padding-left: 1.5em;
            position: relative;
        }

        .description::before {
            content: "?";
            position: absolute;
            left: 0.2em;
            color: var(--link-color);
            font-weight: bold;
        }

        .info-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 20px;
            margin: 20px 0;
        }
        
        .info-card {
            background: var(--info-block-bg);
            border: 1px solid var(--info-block-border);
            border-radius: 16px;
            padding: 20px;
            text-align: center;
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }
        
        .info-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 20px rgba(0,0,0,0.05);
        }

        .info-card .info-label {
            display: block;
            text-transform: uppercase;
            font-size: 0.75rem;
            letter-spacing: 0.1em;
            color: var(--text-secondary);
            margin-bottom: 8px;
            font-weight: 600;
        }
        
        .info-card .info-value {
            font-size: 1.1rem;
            font-weight: 700;
            color: var(--text-primary);
            word-break: break-all;
        }
        
        .info-card .info-value a {
            font-family: 'Outfit', monospace;
            font-weight: 600;
        }
        
        a {
            color: var(--link-color);
            text-decoration: none;
            position: relative;
            transition: all 0.3s ease;
            font-weight: 500;
        }
        
        a::after {
            content: '';
            position: absolute;
            bottom: -2px;
            left: 0;
            width: 0;
            height: 2px;
            background: var(--link-color);
            transition: width 0.3s ease;
        }
        
        a:hover::after {
            width: 100%;
        }
        
        a:hover {
            color: var(--link-hover);
        }
        
        .footer {
            margin-top: 30px;
            text-align: center;
            color: var(--text-secondary);
            font-size: 0.85em;
            opacity: 0.6;
        }
        
        .footer a {
            font-weight: 400;
        }
        
        @media (max-width: 600px) {
            .container { padding: 30px 20px; }
            h1 { font-size: 2.2em; }
            .info-grid { grid-template-columns: 1fr; gap: 12px; }
            .section { padding: 15px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>SubConverter-Extended</h1>
            <p class="subtitle">A Modern Evolution of Subconverter</p>
        </header>

        <div class="info-grid">
            <div class="info-card">
                <span class="info-label">Version</span>
                <div class="info-value">%s</div>
            </div>
            <div class="info-card">
                <span class="info-label">Build</span>
                <div class="info-value">%s</div>
            </div>
            <div class="info-card">
                <span class="info-label">Build Date</span>
                <div class="info-value">%s</div>
            </div>
        </div>
        
        <div class="section">
            <div class="section-title">? Overview</div>
            <p class="description">An enhanced implementation of subconverter, aligned with the <a href="https://github.com/MetaCubeX/mihomo/tree/Meta" target="_blank">Mihomo</a> <a href="https://wiki.metacubex.one/config/" target="_blank">configuration</a>.</p>
            <p class="description">Primarily for <a href="https://github.com/vernesong/OpenClash" target="_blank">OpenClash</a>, while compatible with other Clash clients.</p>
            <p class="description">Dedicated companion backend for the <a href="https://github.com/Aethersailor/Custom_OpenClash_Rules" target="_blank">Custom_OpenClash_Rules</a> project.</p>
        </div>

        <div class="section">
            <div class="section-title">?? Lineage</div>
            <p class="description">Originated and enhanced from: <a href="https://github.com/asdlokj1qpi233/subconverter" target="_blank">subconverter</a></p>
            <p class="description">Modified and evolved by: <a href="https://github.com/Aethersailor" target="_blank">Aethersailor</a></p>
        </div>

        <div class="footer">
            Source Code: <a href="https://github.com/Aethersailor/SubConverter-Extended" target="_blank">GitHub</a> • 
            License: <a href="https://www.gnu.org/licenses/gpl-3.0.html" target="_blank">GPL-3.0</a>
        </div>
    </div>
</body>
</html>`
