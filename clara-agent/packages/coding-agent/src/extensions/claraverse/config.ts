import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { getAgentDir } from "../../config.ts";

/**
 * Clara Agent's own config file, separate from Pi's settings.json so it
 * survives upstream settings-schema changes.
 *
 *   ~/.clara-agent/agent/claraverse.json  ->  { "url": "http://localhost:3000" }
 *
 * Resolution order for the ClaraVerse server URL:
 *   1. CLARAVERSE_URL env var (highest — lets you override per-invocation)
 *   2. the url stored on the saved credential (written at login time)
 *   3. this config file
 *   4. http://localhost:3000
 */

export const DEFAULT_CLARAVERSE_URL = "http://localhost:3000";

export interface ClaraverseConfig {
	url?: string;
}

export function claraverseConfigPath(): string {
	return join(getAgentDir(), "claraverse.json");
}

export function normalizeServerUrl(value: string): string {
	return value.trim().replace(/\/+$/, "");
}

export function readClaraverseConfig(): ClaraverseConfig {
	try {
		const path = claraverseConfigPath();
		if (!existsSync(path)) return {};
		const parsed: unknown = JSON.parse(readFileSync(path, "utf8"));
		if (!parsed || typeof parsed !== "object") return {};
		const url = (parsed as ClaraverseConfig).url;
		return typeof url === "string" && url.trim() ? { url: normalizeServerUrl(url) } : {};
	} catch {
		// A malformed config shouldn't block startup; fall back to defaults.
		return {};
	}
}

export function writeClaraverseConfig(config: ClaraverseConfig): void {
	try {
		const path = claraverseConfigPath();
		mkdirSync(dirname(path), { recursive: true });
		writeFileSync(path, `${JSON.stringify(config, null, 2)}\n`, { mode: 0o600 });
	} catch {
		// Persisting is a convenience, not a requirement — the env var and
		// the credential-stored url both still work without it.
	}
}

/** Server URL to use when no credential has been stored yet (i.e. at login). */
export function resolveServerUrl(): string {
	const fromEnv = process.env.CLARAVERSE_URL?.trim();
	if (fromEnv) return normalizeServerUrl(fromEnv);
	const fromConfig = readClaraverseConfig().url;
	if (fromConfig) return fromConfig;
	return DEFAULT_CLARAVERSE_URL;
}

/**
 * Server URL for an already-authenticated session. The url captured at login
 * wins over the config file so a credential keeps talking to the instance it
 * was actually issued by, but an explicit env var still overrides both.
 */
export function resolveServerUrlForCredential(credentialUrl: unknown): string {
	const fromEnv = process.env.CLARAVERSE_URL?.trim();
	if (fromEnv) return normalizeServerUrl(fromEnv);
	if (typeof credentialUrl === "string" && credentialUrl.trim()) return normalizeServerUrl(credentialUrl);
	return resolveServerUrl();
}
