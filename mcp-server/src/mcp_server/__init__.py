import os
from fastmcp import FastMCP
from google import genai

# Initialize the FastMCP server
mcp = FastMCP("BVR Companion MCP Server")

@mcp.tool()
def sync_gemini_command(command: str) -> str:
    """
    Sync a Google Gemini command locally by sending it to Gemini and returning the response.
    
    Args:
        command: The prompt or command to send to Gemini.
    """
    api_key = os.environ.get("GEMINI_API_KEY")
    if not api_key:
        return "Error: GEMINI_API_KEY environment variable is not set in bvr.json."
        
    try:
        client = genai.Client(api_key=api_key)
        response = client.models.generate_content(
            model='gemini-2.5-flash',
            contents=command
        )
        return f"Gemini Synced Response:\n{response.text}"
    except Exception as e:
        return f"Failed to sync with Gemini: {str(e)}"

def main():
    """Entry point for the MCP server."""
    # Run the server over SSE so the mobile simulator can connect
    mcp.run(transport='sse', host='0.0.0.0', port=8000)
