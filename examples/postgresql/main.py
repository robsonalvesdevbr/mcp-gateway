import os

from mcp import ClientSession
from mcp.client.streamable_http import streamablehttp_client


async def main():
    async with streamablehttp_client(os.getenv("MCP_HOST")) as (
        read_stream,
        write_stream,
        _,
    ):
        async with ClientSession(read_stream, write_stream) as session:
            await session.initialize()
            result = await session.call_tool(
                "query", {"sql": "SELECT datname FROM pg_database;"}
            )
            print(result.content[0].text)


if __name__ == "__main__":
    import asyncio

    asyncio.run(main())
