import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  ListToolsRequestSchema,
  CallToolRequestSchema,
  ErrorCode,
  McpError,
} from "@modelcontextprotocol/sdk/types.js";
import axios from "axios";

// Environment variables provided by Go backend
const SHOPIFY_STORE_URL = process.env.SHOPIFY_STORE_URL;
const SHOPIFY_ACCESS_TOKEN = process.env.SHOPIFY_ACCESS_TOKEN;
const SHOPIFY_API_VERSION = process.env.SHOPIFY_API_VERSION || "2024-01";

if (!SHOPIFY_STORE_URL || !SHOPIFY_ACCESS_TOKEN) {
  console.error("Missing required environment variables: SHOPIFY_STORE_URL or SHOPIFY_ACCESS_TOKEN");
  process.exit(1);
}

const shopifyClient = axios.create({
  baseURL: `https://${SHOPIFY_STORE_URL}/admin/api/${SHOPIFY_API_VERSION}`,
  headers: {
    "X-Shopify-Access-Token": SHOPIFY_ACCESS_TOKEN,
    "Content-Type": "application/json",
  },
});

const server = new Server(
  {
    name: "mi-tech-shopify-mcp",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

/**
 * Tool Definitions
 */
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [
    {
      name: "search_products",
      description: "Search for products in the Shopify store by title or keyword.",
      inputSchema: {
        type: "object",
        properties: {
          query: { type: "string", description: "Product title or keyword to search for" },
          limit: { type: "number", description: "Number of products to return (default 5)", default: 5 },
        },
        required: ["query"],
      },
    },
    {
      name: "get_order_details",
      description: "Get details of a specific order by order number (e.g., #1001).",
      inputSchema: {
        type: "object",
        properties: {
          order_number: { type: "string", description: "The order number including the hash prefix" },
        },
        required: ["order_number"],
      },
    },
    {
      name: "get_fulfillment_tracking",
      description: "Get fulfillment and tracking details for an order ID.",
      inputSchema: {
        type: "object",
        properties: {
          order_id: { type: "string", description: "The internal Shopify order ID" },
        },
        required: ["order_id"],
      },
    },
  ],
}));

/**
 * Tool Implementations
 */
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    switch (name) {
      case "search_products": {
        const response = await shopifyClient.get("/products.json", {
          params: { title: args.query, limit: args.limit || 5 },
        });
        return {
          content: [
            {
              type: "text",
              text: JSON.stringify(response.data.products, null, 2),
            },
          ],
        };
      }

      case "get_order_details": {
        // Shopify REST API uses numeric ID, so we first need to find the order by name
        const searchResponse = await shopifyClient.get("/orders.json", {
          params: { name: args.order_number, status: "any" },
        });

        const order = searchResponse.data.orders?.[0];
        if (!order) {
          throw new McpError(ErrorCode.InvalidRequest, `Order ${args.order_number} not found.`);
        }

        return {
          content: [
            {
              type: "text",
              text: JSON.stringify(order, null, 2),
            },
          ],
        };
      }

      case "get_fulfillment_tracking": {
        const response = await shopifyClient.get(`/orders/${args.order_id}/fulfillments.json`);
        return {
          content: [
            {
              type: "text",
              text: JSON.stringify(response.data.fulfillments, null, 2),
            },
          ],
        };
      }

      default:
        throw new McpError(ErrorCode.MethodNotFound, `Unknown tool: ${name}`);
    }
  } catch (error) {
    if (axios.isAxiosError(error)) {
      return {
        content: [
          {
            type: "text",
            text: `Shopify API Error: ${error.response?.data?.errors || error.message}`,
          },
        ],
        isError: true,
      };
    }
    throw error;
  }
});

/**
 * Transport Setup
 */
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("Shopify MCP Server running on stdio");
}

main().catch((error) => {
  console.error("Server error:", error);
  process.exit(1);
});
