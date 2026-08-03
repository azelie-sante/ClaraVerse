import type { InlineExtension } from "../core/extensions/types.ts";
import claraverseExtension from "./claraverse/index.ts";
import llamaExtension from "./llama/index.ts";

export const builtInExtensions: InlineExtension[] = [
	{ name: "llama.cpp", factory: llamaExtension, hidden: true },
	{ name: "claraverse", factory: claraverseExtension, hidden: true },
];
