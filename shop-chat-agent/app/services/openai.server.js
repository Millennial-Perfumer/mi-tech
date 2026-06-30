/**
 * OpenAI Service
 * Manages interactions with the OpenAI API, translating Anthropic format to OpenAI format.
 */
import OpenAI from "openai";
import AppConfig from "./config.server.js";
import systemPrompts from "../prompts/prompts.json" with { type: "json" };

/**
 * Creates an OpenAI service instance
 * @param {string} apiKey - OpenAI API key
 * @returns {Object} OpenAI service with methods for interacting with OpenAI API
 */
export function createOpenAIService(apiKey = process.env.OPENAI_API_KEY) {
  // Initialize OpenAI client
  const openai = new OpenAI({ apiKey });

  /**
   * Translates Anthropic messages to OpenAI messages format
   */
  const translateMessagesToOpenAI = (claudeMessages) => {
    const openAIMessages = [];

    for (const msg of claudeMessages) {
      if (msg.role === 'user') {
        if (typeof msg.content === 'string') {
          openAIMessages.push({ role: 'user', content: msg.content });
        } else if (Array.isArray(msg.content)) {
          const toolResults = msg.content.filter(block => block.type === 'tool_result');
          if (toolResults.length > 0) {
            for (const block of toolResults) {
              let contentStr = '';
              if (typeof block.content === 'string') {
                contentStr = block.content;
              } else if (Array.isArray(block.content)) {
                contentStr = block.content.map(c => c.text || JSON.stringify(c)).join('\n');
              } else {
                contentStr = JSON.stringify(block.content);
              }
              openAIMessages.push({
                role: 'tool',
                tool_call_id: block.tool_use_id,
                content: contentStr
              });
            }
          } else {
            const text = msg.content.map(block => block.text || '').join('\n');
            openAIMessages.push({ role: 'user', content: text });
          }
        }
      } else if (msg.role === 'assistant') {
        if (typeof msg.content === 'string') {
          openAIMessages.push({ role: 'assistant', content: msg.content });
        } else if (Array.isArray(msg.content)) {
          const textBlock = msg.content.find(block => block.type === 'text');
          const textContent = textBlock ? textBlock.text : null;

          const toolUseBlocks = msg.content.filter(block => block.type === 'tool_use');
          const toolCalls = toolUseBlocks.map(block => ({
            id: block.id,
            type: 'function',
            function: {
              name: block.name,
              arguments: JSON.stringify(block.input)
            }
          }));

          openAIMessages.push({
            role: 'assistant',
            content: textContent || undefined,
            tool_calls: toolCalls.length > 0 ? toolCalls : undefined
          });
        }
      }
    }

    return openAIMessages;
  };

  /**
   * Translates Anthropic tools to OpenAI tools format
   */
  const translateToolsToOpenAI = (claudeTools) => {
    if (!claudeTools || claudeTools.length === 0) return undefined;
    return claudeTools.map(tool => ({
      type: 'function',
      function: {
        name: tool.name,
        description: tool.description,
        parameters: tool.input_schema
      }
    }));
  };

  /**
   * Streams a conversation with OpenAI
   */
  const streamConversation = async ({
    messages,
    promptType = AppConfig.api.defaultPromptType,
    tools
  }, streamHandlers) => {
    const systemInstruction = getSystemPrompt(promptType);

    const openAIMessages = [
      { role: 'system', content: systemInstruction },
      ...translateMessagesToOpenAI(messages)
    ];

    const openAITools = translateToolsToOpenAI(tools);

    // Determine model
    let model = AppConfig.api.defaultModel;
    if (!model || model.startsWith('claude')) {
      model = 'gpt-5.4-mini';
    }

    const response = await openai.chat.completions.create({
      model: model,
      messages: openAIMessages,
      tools: openAITools,
      stream: true,
      max_completion_tokens: AppConfig.api.maxTokens
    });

    let completeText = "";
    let toolCallsAccumulator = {};

    for await (const chunk of response) {
      const choice = chunk.choices[0];
      if (!choice) continue;

      if (choice.delta?.content) {
        const textDelta = choice.delta.content;
        completeText += textDelta;
        if (streamHandlers.onText) {
          streamHandlers.onText(textDelta);
        }
      }

      if (choice.delta?.tool_calls) {
        for (const tc of choice.delta.tool_calls) {
          const idx = tc.index;
          if (!toolCallsAccumulator[idx]) {
            toolCallsAccumulator[idx] = {
              id: tc.id,
              name: tc.function?.name || "",
              arguments: tc.function?.arguments || ""
            };
          } else {
            if (tc.id) toolCallsAccumulator[idx].id = tc.id;
            if (tc.function?.name) toolCallsAccumulator[idx].name += tc.function.name;
            if (tc.function?.arguments) toolCallsAccumulator[idx].arguments += tc.function.arguments;
          }
        }
      }
    }

    // Process result
    const hasToolCalls = Object.keys(toolCallsAccumulator).length > 0;
    const contentBlocks = [];

    if (completeText) {
      contentBlocks.push({ type: 'text', text: completeText });
    }

    if (hasToolCalls) {
      const toolUses = Object.values(toolCallsAccumulator).map(tc => {
        let parsedArgs = {};
        try {
          parsedArgs = JSON.parse(tc.arguments);
        } catch (e) {
          console.error("Failed to parse tool arguments:", tc.arguments);
        }
        return {
          type: 'tool_use',
          id: tc.id,
          name: tc.name,
          input: parsedArgs
        };
      });

      contentBlocks.push(...toolUses);

      const finalMessage = {
        role: 'assistant',
        content: contentBlocks,
        stop_reason: 'tool_use'
      };

      if (streamHandlers.onMessage) {
        streamHandlers.onMessage(finalMessage);
      }

      if (streamHandlers.onToolUse) {
        for (const toolUse of toolUses) {
          await streamHandlers.onToolUse(toolUse);
        }
      }

      return finalMessage;
    } else {
      const finalMessage = {
        role: 'assistant',
        content: contentBlocks,
        stop_reason: 'end_turn'
      };

      if (streamHandlers.onMessage) {
        streamHandlers.onMessage(finalMessage);
      }

      return finalMessage;
    }
  };

  const getSystemPrompt = (promptType) => {
    return systemPrompts.systemPrompts[promptType]?.content ||
      systemPrompts.systemPrompts[AppConfig.api.defaultPromptType].content;
  };

  return {
    streamConversation,
    getSystemPrompt
  };
}

export default {
  createOpenAIService
};
