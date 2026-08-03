import type {
	AuthInteraction,
	Model,
	OAuthAuth,
	OAuthCredential,
	Provider,
	ProviderStreamOptions,
	RefreshModelsContext,
} from "@earendil-works/pi-ai";
import { stream, streamSimple } from "@earendil-works/pi-ai/compat";
import { VERSION } from "../../config.ts";
import { resolveServerUrl, resolveServerUrlForCredential, writeClaraverseConfig } from "./config.ts";
import { pollDeviceCodeFlow } from "./device-code-poll.ts";

export const CLARAVERSE_PROVIDER_ID = "claraverse";

/** Credential extras stored alongside the standard OAuth fields. */
interface ClaraverseCredential extends OAuthCredential {
	deviceId: string;
	/** ClaraVerse instance this credential was issued by. */
	serverUrl?: string;
}

/** Server URL for a stored credential, honoring env override and config file. */
function credentialServerUrl(credential: OAuthCredential | undefined): string {
	return resolveServerUrlForCredential((credential as ClaraverseCredential | undefined)?.serverUrl);
}

async function fetchJson(url: string, init?: RequestInit): Promise<unknown> {
	const response = await fetch(url, init);
	const text = await response.text();
	let data: unknown;
	try {
		data = text ? JSON.parse(text) : undefined;
	} catch {
		data = undefined;
	}
	if (!response.ok) {
		throw new Error(`${response.status} ${response.statusText}${text ? `: ${text}` : ""}`);
	}
	return data;
}

interface DeviceCodeResponse {
	device_code: string;
	user_code: string;
	verification_uri: string;
	verification_uri_complete: string;
	expires_in: number;
	interval: number;
}

interface TokenResponse {
	access_token: string;
	token_type: string;
	expires_in: number;
	refresh_token: string;
	device_id: string;
	user: { id: string; email: string };
}

interface TokenErrorResponse {
	error: string;
	error_description?: string;
	interval?: number;
}

function isTokenResponse(value: unknown): value is TokenResponse {
	return !!value && typeof value === "object" && typeof (value as TokenResponse).access_token === "string";
}

async function startDeviceFlow(serverUrl: string): Promise<DeviceCodeResponse> {
	const data = await fetchJson(`${serverUrl}/api/device/code`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({
			client_id: "clara-agent",
			client_version: VERSION,
			platform: process.platform,
		}),
	});
	if (!data || typeof data !== "object" || typeof (data as DeviceCodeResponse).device_code !== "string") {
		throw new Error("Invalid device code response from ClaraVerse");
	}
	return data as DeviceCodeResponse;
}

async function pollForToken(serverUrl: string, deviceCode: string, signal?: AbortSignal): Promise<TokenResponse> {
	return pollDeviceCodeFlow<TokenResponse>({
		waitBeforeFirstPoll: true,
		signal,
		poll: async () => {
			const url = `${serverUrl}/api/device/token?device_code=${encodeURIComponent(deviceCode)}&client_id=clara-agent`;
			const response = await fetch(url);
			const raw = (await response.json().catch(() => undefined)) as unknown;
			if (isTokenResponse(raw)) {
				return { status: "complete", value: raw };
			}
			const err = raw as TokenErrorResponse | undefined;
			if (err?.error === "authorization_pending") return { status: "pending" };
			if (err?.error === "slow_down") return { status: "slow_down", intervalSeconds: err.interval };
			return { status: "failed", message: err?.error_description || err?.error || "Device flow failed" };
		},
	});
}

async function refreshClaraverseToken(
	serverUrl: string,
	refreshToken: string,
	deviceId: string,
): Promise<TokenResponse> {
	const data = await fetchJson(`${serverUrl}/api/devices/refresh-token`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ refresh_token: refreshToken, device_id: deviceId }),
	});
	if (!isTokenResponse(data)) {
		throw new Error("Invalid refresh response from ClaraVerse");
	}
	return data;
}

interface ClaraverseModelInfo {
	id: string;
	name: string;
	display_name?: string;
	context_length?: number;
	supports_tools?: boolean;
	supports_vision?: boolean;
}

interface ProviderConfigResponse {
	base_url: string;
	api_key?: string;
	default_model?: string;
	models?: ClaraverseModelInfo[];
}

async function fetchProviderConfig(serverUrl: string, accessToken: string): Promise<ProviderConfigResponse> {
	let data: unknown;
	try {
		data = await fetchJson(`${serverUrl}/api/agent/provider-config`, {
			headers: { Authorization: `Bearer ${accessToken}` },
		});
	} catch (error) {
		const message = error instanceof Error ? error.message : String(error);
		// This endpoint was added for Clara Agent, so an older ClaraVerse
		// build won't have it. Say so plainly instead of surfacing a bare 404.
		if (message.includes("404")) {
			throw new Error(
				`ClaraVerse at ${serverUrl} does not support Clara Agent yet (no /api/agent/provider-config). ` +
					`Update that instance, or point Clara Agent at one that has it with CLARAVERSE_URL.`,
			);
		}
		throw error;
	}
	if (!data || typeof data !== "object" || typeof (data as ProviderConfigResponse).base_url !== "string") {
		throw new Error("Invalid provider config response from ClaraVerse. Is a default model configured?");
	}
	return data as ProviderConfigResponse;
}

function toCredential(token: TokenResponse, serverUrl: string): ClaraverseCredential {
	return {
		type: "oauth",
		access: token.access_token,
		refresh: token.refresh_token,
		expires: Date.now() + token.expires_in * 1000 - 5 * 60_000,
		deviceId: token.device_id,
		serverUrl,
	};
}

async function loginClaraverse(interaction: AuthInteraction): Promise<OAuthCredential> {
	const serverUrl = resolveServerUrl();
	const device = await startDeviceFlow(serverUrl);
	interaction.notify({
		type: "device_code",
		userCode: device.user_code,
		verificationUri: device.verification_uri_complete || device.verification_uri,
		intervalSeconds: device.interval,
		expiresInSeconds: device.expires_in,
	});
	const token = await pollForToken(serverUrl, device.device_code, interaction.signal);
	// Remember which instance authorized us, so later runs don't need the env
	// var and a re-login against a different instance updates it.
	writeClaraverseConfig({ url: serverUrl });
	return toCredential(token, serverUrl);
}

/**
 * Local-model setups often identify a model by its full .gguf path, which
 * makes the /model picker unreadable. Show just the model file's base name
 * while keeping `id` exactly as the provider expects it.
 */
function displayLabel(info: ClaraverseModelInfo): string {
	const raw = info.display_name || info.name || info.id;
	if (!raw.includes("/") && !raw.endsWith(".gguf")) return raw;
	const base = raw.split(/[/\\]/).pop() ?? raw;
	return base.replace(/\.gguf$/i, "") || raw;
}

function toAgentModel(info: ClaraverseModelInfo, baseUrl: string): Model<"openai-completions"> {
	// ClaraVerse reports the real context length per model; only fall back to
	// a conservative default when it's missing or nonsensical.
	const contextWindow = info.context_length && info.context_length > 0 ? info.context_length : 128000;
	return {
		id: info.id,
		name: displayLabel(info),
		api: "openai-completions",
		provider: CLARAVERSE_PROVIDER_ID,
		baseUrl,
		reasoning: false,
		input: info.supports_vision ? ["text", "image"] : ["text"],
		// Models reached through ClaraVerse are billed (or not) by the
		// upstream provider, not metered here.
		cost: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0 },
		contextWindow,
		maxTokens: contextWindow,
		compat: {
			supportsStore: false,
			supportsDeveloperRole: false,
			supportsReasoningEffort: false,
			supportsUsageInStreaming: true,
			supportsStrictMode: false,
			maxTokensField: "max_tokens",
		},
	};
}

export interface ClaraverseProviderController {
	provider: Provider<"openai-completions">;
}

/**
 * The ClaraVerse account provider. Auth is a standard RFC 8628 device flow
 * against the connected ClaraVerse instance (same mechanism the Devices
 * settings page and other connected clients use). `toAuth()`/`refreshModels()`
 * then call ClaraVerse's `GET /api/agent/provider-config`, which resolves the
 * account's actually-configured model provider and hands back its real
 * base URL and API key, so Clara Agent talks to that provider directly.
 *
 * Known limitation: this assumes the configured provider speaks the
 * OpenAI-compatible chat completions format (true for local llama.cpp,
 * Ollama, OpenAI itself, and any "OpenAI-compatible endpoint" — which
 * covers ClaraVerse's default/common case). A provider configured with a
 * different wire format (raw Anthropic Messages API, raw Gemini, etc.)
 * is not yet mapped correctly here.
 */
export function createClaraverseProvider(): ClaraverseProviderController {
	let models: Model<"openai-completions">[] = [];

	const oauth: OAuthAuth = {
		name: "ClaraVerse Account",
		login: loginClaraverse,
		refresh: async (credential) => {
			const deviceId = (credential as ClaraverseCredential).deviceId;
			if (!deviceId) {
				throw new Error("Missing device id on stored ClaraVerse credential. Please log in again.");
			}
			const serverUrl = credentialServerUrl(credential);
			return toCredential(await refreshClaraverseToken(serverUrl, credential.refresh, deviceId), serverUrl);
		},
		toAuth: async (credential) => {
			const config = await fetchProviderConfig(credentialServerUrl(credential), credential.access);
			return { apiKey: config.api_key ?? "local", baseUrl: config.base_url };
		},
	};

	const provider: Provider<"openai-completions"> = {
		id: CLARAVERSE_PROVIDER_ID,
		name: "ClaraVerse Account",
		baseUrl: resolveServerUrl(),
		auth: { oauth },
		getModels: () => models,
		refreshModels: async (context: RefreshModelsContext): Promise<void> => {
			const stored = await context.store.read();
			if (stored) {
				models = stored.models.filter(
					(model): model is Model<"openai-completions"> =>
						model.provider === CLARAVERSE_PROVIDER_ID && model.api === "openai-completions",
				);
			}

			if (!context.allowNetwork || context.signal?.aborted || context.credential?.type !== "oauth") return;
			const credential = context.credential as ClaraverseCredential;
			const config = await fetchProviderConfig(credentialServerUrl(credential), credential.access);
			const catalog: ClaraverseModelInfo[] = config.models?.length
				? config.models
				: config.default_model
					? [{ id: config.default_model, name: config.default_model }]
					: [];
			models = catalog.map((info) => toAgentModel(info, config.base_url));
			if (!context.signal?.aborted) await context.store.write({ models, checkedAt: Date.now() });
		},
		stream: (model, ctx, options) => stream(model, ctx, options as ProviderStreamOptions | undefined),
		streamSimple: (model, ctx, options) => streamSimple(model, ctx, options),
	};

	return { provider };
}
