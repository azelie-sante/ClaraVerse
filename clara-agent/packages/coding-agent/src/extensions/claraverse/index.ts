import type { ExtensionAPI } from "../../core/extensions/types.ts";
import { createClaraverseProvider } from "./provider.ts";

export default function claraverseExtension(pi: ExtensionAPI): void {
	const { provider } = createClaraverseProvider();
	pi.registerProvider(provider);
}
