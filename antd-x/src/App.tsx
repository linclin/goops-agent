import {
  CloudUploadOutlined,
  DeleteOutlined,
  EditOutlined,
  EllipsisOutlined,
  PaperClipOutlined,
  QuestionCircleOutlined,
  ShareAltOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import type { BubbleListProps } from '@ant-design/x';
import {
  Actions,
  Attachments,
  Bubble,
  Conversations,
  Prompts,
  Sender,
  Welcome,
  XProvider,
} from '@ant-design/x';
import { XMarkdown } from '@ant-design/x-markdown';
import { XRequest } from '@ant-design/x-sdk';
import { Avatar, Button, Flex, type GetProp, message, Pagination, Space } from 'antd';
import { createStyles } from 'antd-style';
import dayjs from 'dayjs';
import React, { useRef, useState } from 'react';
import { BubbleListRef } from '@ant-design/x/es/bubble';
import locale from './pages/_utils/local';

// ==================== Style ====================
const useStyle = createStyles(({ token, css }) => {
  return {
    layout: css`
      width: 100%;
      height: 100vh;
      display: flex;
      background: ${token.colorBgContainer};
      font-family: AlibabaPuHuiTi, ${token.fontFamily}, sans-serif;
    `,
    side: css`
      background: ${token.colorBgLayout}80;
      width: 280px;
      height: 100%;
      display: flex;
      flex-direction: column;
      padding: 0 12px;
      box-sizing: border-box;
    `,
    logo: css`
      display: flex;
      align-items: center;
      justify-content: start;
      padding: 0 24px;
      box-sizing: border-box;
      gap: 8px;
      margin: 24px 0;

      span {
        font-weight: bold;
        color: ${token.colorText};
        font-size: 16px;
      }
    `,
    conversations: css`
      overflow-y: auto;
      margin-top: 12px;
      padding: 0;
      flex: 1;
      .ant-conversations-list {
        padding-inline-start: 0;
      }
    `,
    sideFooter: css`
      border-top: 1px solid ${token.colorBorderSecondary};
      height: 40px;
      display: flex;
      align-items: center;
      justify-content: space-between;
    `,
    chat: css`
      height: 100%;
      width: calc(100% - 280px);
      box-sizing: border-box;
      display: flex;
      flex-direction: column;
      padding-block: ${token.paddingLG}px;
      justify-content: space-between;
    `,
    chatPrompt: css`
      .ant-prompts-label {
        color: #000000e0 !important;
      }
      .ant-prompts-desc {
        color: #000000a6 !important;
        width: 100%;
      }
      .ant-prompts-icon {
        color: #000000a6 !important;
      }
    `,
    chatList: css`
      display: flex;
      height: calc(100% - 120px);
      flex-direction: column;
      align-items: center;
      width: 100%;
    `,
    placeholder: css`
      padding-top: 32px;
      width: 100%;
      padding-inline: ${token.paddingLG}px;
      box-sizing: border-box;
    `,
    sender: css`
      width: 100%;
      max-width: 840px;
    `,
    senderPrompt: css`
      width: 100%;
      max-width: 840px;
      margin: 0 auto;
      color: ${token.colorText};
    `,
  };
});

// ==================== Static Config ====================
const DEFAULT_CONVERSATIONS_ITEMS = [
  {
    key: 'default-0',
    label: locale.whatIsAntDesignX,
    group: locale.today,
  },
  {
    key: 'default-1',
    label: locale.howToQuicklyInstallAndImportComponents,
    group: locale.today,
  },
  {
    key: 'default-2',
    label: locale.newAgiHybridInterface,
    group: locale.yesterday,
  },
];

const HOT_TOPICS = {
  key: '1',
  label: locale.hotTopics,
  children: [
    {
      key: '1-1',
      description: locale.whatComponentsAreInAntDesignX,
      icon: <span style={{ color: '#f93a4a', fontWeight: 700 }}>1</span>,
    },
    {
      key: '1-2',
      description: locale.newAgiHybridInterface,
      icon: <span style={{ color: '#ff6565', fontWeight: 700 }}>2</span>,
    },
    {
      key: '1-3',
      description: locale.whatComponentsAreInAntDesignX,
      icon: <span style={{ color: '#ff8f1f', fontWeight: 700 }}>3</span>,
    },
    {
      key: '1-4',
      description: locale.comeAndDiscoverNewDesignParadigm,
      icon: <span style={{ color: '#00000040', fontWeight: 700 }}>4</span>,
    },
    {
      key: '1-5',
      description: locale.howToQuicklyInstallAndImportComponents,
      icon: <span style={{ color: '#00000040', fontWeight: 700 }}>5</span>,
    },
  ],
};

const SENDER_PROMPTS: GetProp<typeof Prompts, 'items'> = [
  {
    key: '1',
    description: locale.upgrades,
  },
  {
    key: '2',
    description: locale.components,
  },
  {
    key: '3',
    description: locale.richGuide,
  },
  {
    key: '4',
    description: locale.installationIntroduction,
  },
];

// ==================== Type ====================
interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  status?: 'loading' | 'success' | 'error';
}

// ==================== Sub Component ====================
const Footer: React.FC<{
  id?: string | number;
  content: string;
  status?: string;
}> = ({ id, content, status }) => {
  const Items = [
    {
      key: 'pagination',
      actionRender: <Pagination simple total={1} pageSize={1} />,
    },
    {
      key: 'retry',
      label: locale.retry,
      icon: <SyncOutlined />,
    },
    {
      key: 'copy',
      actionRender: <Actions.Copy text={content} />,
    },
    {
      key: 'audio',
      actionRender: (
        <Actions.Audio
          onClick={() => {
            message.info(locale.isMock);
          }}
        />
      ),
    },
  ];
  return status !== 'updating' && status !== 'loading' ? (
    <div style={{ display: 'flex' }}>{id && <Actions items={Items} />}</div>
  ) : null;
};

// 客户端渲染包装组件
const ClientOnly: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isClient, setIsClient] = React.useState(false);

  React.useEffect(() => {
    setIsClient(true);
  }, []);

  if (!isClient) {
    return null;
  }

  return <>{children}</>;
};

const Independent: React.FC = () => {
  const { styles } = useStyle();

  // ==================== State ====================
  const [conversations, setConversations] = useState(DEFAULT_CONVERSATIONS_ITEMS);
  const [activeConversationKey, setActiveConversationKey] = useState(DEFAULT_CONVERSATIONS_ITEMS[0].key);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isRequesting, setIsRequesting] = useState(false);
  const [messageApi, contextHolder] = message.useMessage();
  const [attachmentsOpen, setAttachmentsOpen] = useState(false);
  const [attachedFiles, setAttachedFiles] = useState<GetProp<typeof Attachments, 'items'>>([]);
  const [inputValue, setInputValue] = useState('');
  const listRef = useRef<BubbleListRef>(null);

  // ==================== XRequest ====================
  const sendMessage = async (content: string) => {
    if (!content.trim()) return;

    // 添加用户消息
    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: content,
      status: 'success',
    };
    setMessages((prev) => [...prev, userMessage]);
    setInputValue('');
    setIsRequesting(true);

    // 添加助手消息（加载中）
    const assistantId = (Date.now() + 1).toString();
    setMessages((prev) => [
      ...prev,
      {
        id: assistantId,
        role: 'assistant',
        content: '',
        status: 'loading',
      },
    ]);

    try {
      let assistantContent = '';

      XRequest('http://127.0.0.1:8080/api/v1/agent/chat/sse', {
        params: {
          id: activeConversationKey,
          query: content,
          history: messages.filter(m => m.status === 'success' && m.content).map(m => ({
            role: m.role,
            content: typeof m.content === 'string' ? m.content : JSON.stringify(m.content),
          })),
        },
        headers: {
          'Authorization': 'Basic ' + btoa('admin:admin'),
        },
        callbacks: {
          onSuccess: (messages) => {
            console.log('onSuccess', messages);
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantId ? { ...m, content: assistantContent, status: 'success' } : m
              )
            );
          },
          onError: (error) => {
            console.error('onError', error);
            let errorMessage = locale.requestFailed;
            
            if (error instanceof Error) {
              if (error.message.includes('404')) {
                errorMessage = '接口不存在，请检查后端服务是否运行';
              } else if (error.message.includes('401')) {
                errorMessage = '认证失败，请检查用户名和密码';
              } else if (error.message.includes('403')) {
                errorMessage = '权限不足，无法访问该接口';
              } else if (error.message.includes('network')) {
                errorMessage = '网络连接失败，请检查网络设置';
              } else {
                errorMessage = `请求失败: ${error.message}`;
              }
            }
            
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantId
                  ? { ...m, content: errorMessage, status: 'error' }
                  : m
              )
            );
            messageApi.error(errorMessage);
          },
          onUpdate: (msg) => {
            console.log('onUpdate raw:', msg, 'type:', typeof msg);
            // 解析 SSE 消息，提取 data 字段
            let text = '';
            
            try {
              // 将 msg 转换为字符串（如果是对象则转为 JSON 字符串）
              const msgStr = typeof msg === 'object' ? JSON.stringify(msg) : String(msg);
              console.log('msgStr:', msgStr);
              
              // 尝试解析 JSON
              const parsed = JSON.parse(msgStr);
              console.log('parsed:', parsed);
              
              // 提取 data 字段（data 字段可能是字符串或 JSON 字符串）
              if (parsed && parsed.data) {
                const dataValue = parsed.data;
                console.log('dataValue:', dataValue, 'type:', typeof dataValue);
                
                // 如果 data 是 JSON 字符串，再次解析
                if (typeof dataValue === 'string' && dataValue.startsWith('{')) {
                  try {
                    const innerParsed = JSON.parse(dataValue);
                    if (innerParsed.data) {
                      text = innerParsed.data;
                    } else {
                      text = dataValue;
                    }
                  } catch (e) {
                    // data 不是 JSON 字符串，直接使用
                    text = dataValue;
                  }
                } else {
                  text = dataValue;
                }
                console.log('extracted text:', text);
              }
            } catch (e) {
              console.error('Parse error:', e);
              // 解析失败，尝试直接使用
              text = String(msg);
            }
            
            if (text && text !== '[object Object]') {
              assistantContent += text;
              setMessages((prev) =>
                prev.map((m) =>
                  m.id === assistantId
                    ? { ...m, content: assistantContent, status: 'success' }
                    : m
                )
              );
            }
          },
        },
      });
    } catch (error) {
      console.error('Request failed:', error);
      let errorMessage = locale.requestFailed;
      
      if (error instanceof Error) {
        if (error.message.includes('404')) {
          errorMessage = '接口不存在，请检查后端服务是否运行';
        } else if (error.message.includes('401')) {
          errorMessage = '认证失败，请检查用户名和密码';
        } else if (error.message.includes('403')) {
          errorMessage = '权限不足，无法访问该接口';
        } else if (error.message.includes('network')) {
          errorMessage = '网络连接失败，请检查网络设置';
        } else {
          errorMessage = `请求失败: ${error.message}`;
        }
      }
      
      setMessages((prev) =>
        prev.map((m) =>
          m.id === assistantId
            ? { ...m, content: errorMessage, status: 'error' }
            : m
        )
      );
      messageApi.error(errorMessage);
    } finally {
      setIsRequesting(false);
    }
  };

  // ==================== Event ====================
  const onSubmit = async (val: string) => {
    await sendMessage(val);
    listRef.current?.scrollTo({ top: 'bottom' });
  };

  const addConversation = (conversation: { key: string; label: string; group: string }) => {
    setConversations((prev) => [...prev, conversation]);
    setMessages([]);
  };

  // ==================== Nodes ====================
  const chatSide = (
    <div className={styles.side}>
      {/* Logo */}
      <div className={styles.logo}>
        <img
          src="https://mdn.alipayobjects.com/huamei_iwk9zp/afts/img/A*eco6RrQhxbMAAAAAAAAAAAAADgCCAQ/original"
          draggable={false}
          alt="logo"
          width={24}
          height={24}
        />
        <span>Ant Design X</span>
      </div>

      {/* 会话管理 */}
      <Conversations
        creation={{
          onClick: () => {
            if (messages.length === 0) {
              messageApi.error(locale.itIsNowANewConversation);
              return;
            }
            const now = dayjs().valueOf().toString();
            addConversation({
              key: now,
              label: `${locale.newConversation} ${conversations.length + 1}`,
              group: locale.today,
            });
            setActiveConversationKey(now);
          },
        }}
        items={conversations.map(({ key, label, ...other }) => ({
          key,
          label: key === activeConversationKey ? `[${locale.curConversation}]${label}` : label,
          ...other,
        }))}
        className={styles.conversations}
        activeKey={activeConversationKey}
        onActiveChange={(key) => {
          setActiveConversationKey(key);
          setMessages([]);
        }}
        groupable
        styles={{ item: { padding: '0 8px' } }}
        menu={(conversation) => ({
          items: [
            {
              label: locale.rename,
              key: 'rename',
              icon: <EditOutlined />,
            },
            {
              label: locale.delete,
              key: 'delete',
              icon: <DeleteOutlined />,
              danger: true,
              onClick: () => {
                const newList = conversations.filter((item) => item.key !== conversation.key);
                const newKey = newList?.[0]?.key;
                setConversations(newList);
                if (conversation.key === activeConversationKey) {
                  setActiveConversationKey(newKey);
                  setMessages([]);
                }
              },
            },
          ],
        })}
      />

      <div className={styles.sideFooter}>
        <Avatar size={24} />
        <Button type="text" icon={<QuestionCircleOutlined />} />
      </div>
    </div>
  );

  const getRole = (): BubbleListProps['role'] => ({
    assistant: {
      placement: 'start',
      avatar: {
        icon: <img src="https://mdn.alipayobjects.com/huamei_iwk9zp/afts/img/A*s5sNRo5LjfQAAAAAAAAAAAAADgCCAQ/fmt.webp" width={32} />,
      },
      footer: (content, { status, key }) => (
        <Footer content={content} status={status} id={key as string} />
      ),
    },
    user: { placement: 'end' },
  });

  const chatList = (
    <div className={styles.chatList}>
      {messages?.length ? (
        <Bubble.List
          ref={listRef}
          items={messages?.map((i) => ({
            key: i.id,
            role: i.role,
            content: i.role === 'assistant' ? <XMarkdown content={i.content} /> : i.content,
            status: i.status,
            loading: i.status === 'loading',
          }))}
          styles={{
            root: {
              maxWidth: 940,
            },
          }}
          role={getRole()}
        />
      ) : (
        <Flex
          vertical
          style={{
            maxWidth: 840,
          }}
          gap={16}
          align="center"
          className={styles.placeholder}
        >
          <Welcome
            style={{
              width: '100%',
            }}
            variant="borderless"
            icon="https://mdn.alipayobjects.com/huamei_iwk9zp/afts/img/A*s5sNRo5LjfQAAAAAAAAAAAAADgCCAQ/fmt.webp"
            title={locale.welcome}
            description={locale.welcomeDescription}
            extra={
              <Space>
                <Button icon={<ShareAltOutlined />} />
                <Button icon={<EllipsisOutlined />} />
              </Space>
            }
          />
          <Flex
            gap={16}
            justify="center"
            style={{
              width: '100%',
            }}
          >
            <Prompts
              items={[HOT_TOPICS]}
              styles={{
                list: { height: '100%' },
                item: {
                  flex: 1,
                  backgroundImage: 'linear-gradient(123deg, #e5f4ff 0%, #efe7ff 100%)',
                  borderRadius: 12,
                  border: 'none',
                },
                subItem: { padding: 0, background: 'transparent' },
              }}
              onItemClick={(info) => {
                onSubmit(info.data.description as string);
              }}
              className={styles.chatPrompt}
            />
          </Flex>
        </Flex>
      )}
    </div>
  );

  const senderHeader = (
    <Sender.Header
      title={locale.uploadFile}
      open={attachmentsOpen}
      onOpenChange={setAttachmentsOpen}
      styles={{ content: { padding: 0 } }}
    >
      <Attachments
        beforeUpload={() => false}
        items={attachedFiles}
        onChange={(info) => setAttachedFiles(info.fileList)}
        placeholder={(type) =>
          type === 'drop'
            ? { title: locale.dropFileHere }
            : {
                icon: <CloudUploadOutlined />,
                title: locale.uploadFiles,
                description: locale.clickOrDragFilesToUpload,
              }
        }
      />
    </Sender.Header>
  );

  const chatSender = (
    <Flex
      vertical
      gap={12}
      align="center"
      style={{
        marginInline: 24,
      }}
    >
      {/* 提示词 */}
      {!attachmentsOpen && (
        <Prompts
          items={SENDER_PROMPTS}
          onItemClick={(info) => {
            onSubmit(info.data.description as string);
          }}
          styles={{
            item: { padding: '6px 12px' },
          }}
          className={styles.senderPrompt}
        />
      )}
      {/* 输入框 */}
      <Sender
        value={inputValue}
        header={senderHeader}
        onSubmit={() => {
          onSubmit(inputValue);
        }}
        onChange={setInputValue}
        onCancel={() => {
          setIsRequesting(false);
        }}
        prefix={
          <Button
            type="text"
            icon={<PaperClipOutlined style={{ fontSize: 18 }} />}
            onClick={() => setAttachmentsOpen(!attachmentsOpen)}
          />
        }
        loading={isRequesting}
        className={styles.sender}
        allowSpeech
        placeholder={locale.askOrInputUseSkills}
      />
    </Flex>
  );

  // ==================== Render =================
  return (
    <XProvider locale={locale}>
      {contextHolder}
      <div className={styles.layout}>
        {chatSide}
        <div className={styles.chat}>
          {chatList}
          {chatSender}
        </div>
      </div>
    </XProvider>
  );
};

const App: React.FC = () => {
  return (
    <ClientOnly>
      <Independent />
    </ClientOnly>
  );
};

export default App;
