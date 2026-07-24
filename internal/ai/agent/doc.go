// Package agent implements a multi-turn conversational LLM agent for LifeOS.
//
// ChatModel adapters: *ollama.Client and *openaiapi.Client already expose
// Chat / ChatJSON with matching signatures and therefore satisfy ai.ChatModel
// without wrappers.
package agent
