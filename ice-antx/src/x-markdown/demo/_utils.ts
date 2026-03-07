import { theme } from 'antd';
import React from 'react';

const splitIntoChunks = (str: string, chunkSize: number) => {
  const chunks = [];
  for (let i = 0; i < str.length; i += chunkSize) {
    chunks.push(str.slice(i, i + chunkSize));
  }
  return chunks;
};

export const mockFetch = async (fullContent: string, onFinish?: () => void) => {
  const chunks = splitIntoChunks(fullContent, 7);
  const response = new Response(
    new ReadableStream({
      async start(controller) {
        try {
          await new Promise((resolve) => setTimeout(resolve, 100));
          for (const chunk of chunks) {
            await new Promise((resolve) => setTimeout(resolve, 100));
            if (!controller.desiredSize) {
              // 流已满或关闭，避免写入
              return;
            }
            controller.enqueue(new TextEncoder().encode(chunk));
          }
          onFinish?.();
          controller.close();
        } catch (error) {
          console.log(error);
        }
      },
    }),
    {
      headers: {
        'Content-Type': 'application/x-ndjson',
      },
    },
  );

  return response;
};

export const useMarkdownTheme = () => {
  const token = theme.useToken();
  const [isClient, setIsClient] = React.useState(false);

  React.useEffect(() => {
    setIsClient(true);
  }, []);

  // 在服务器端默认使用亮色模式，避免 hydration 错误
  const isLightMode = React.useMemo(() => {
    if (!isClient) {
      return true;
    }
    return token?.theme?.id === 0;
  }, [token, isClient]);

  const className = React.useMemo(() => {
    return isLightMode ? 'x-markdown-light' : 'x-markdown-dark';
  }, [isLightMode]);

  return [className];
};
