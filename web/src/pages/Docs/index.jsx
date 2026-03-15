import React, { useContext, useMemo } from 'react';
import { Button, Card, Space, Typography } from '@douyinfe/semi-ui';
import { Link } from 'react-router-dom';
import { copy, getSystemName, showSuccess } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { useTranslation } from 'react-i18next';

const { Title, Text, Paragraph } = Typography;

const Docs = () => {
  const { i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const systemName = getSystemName();
  const isChinese = i18n.language.startsWith('zh');
  const serverAddress =
    statusState?.status?.server_address || window.location.origin;
  const apiBaseUrl = `${serverAddress}/v1`;

  const examples = useMemo(
    () => ({
      curl: `curl ${apiBaseUrl}/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'`,
      python: `from openai import OpenAI

client = OpenAI(
    api_key="YOUR_API_KEY",
    base_url="${apiBaseUrl}"
)

resp = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello"}],
)

print(resp.choices[0].message.content)`,
      javascript: `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "YOUR_API_KEY",
  baseURL: "${apiBaseUrl}",
});

const resp = await client.chat.completions.create({
  model: "gpt-4o-mini",
  messages: [{ role: "user", content: "Hello" }],
});

console.log(resp.choices[0].message.content);`,
    }),
    [apiBaseUrl],
  );

  const handleCopy = async (value) => {
    const ok = await copy(value);
    if (ok) showSuccess(isChinese ? '已复制到剪贴板' : 'Copied');
  };

  return (
    <div className='mx-auto max-w-5xl px-4 py-24'>
      <Space vertical spacing='loose' style={{ width: '100%' }}>
        <div>
          <Title heading={2}>
            {isChinese ? `${systemName} 接入文档` : `${systemName} Quick Start`}
          </Title>
          <Paragraph type='secondary'>
            {isChinese
              ? '你充值后创建的 sk- 密钥，就是调用接口时放在 Authorization 里的 Bearer Token。'
              : 'Use your generated sk- key as the Bearer token in Authorization.'}
          </Paragraph>
        </div>

        <Card>
          <Space vertical spacing='medium' style={{ width: '100%' }}>
            <Title heading={4}>{isChinese ? '1. 调用地址' : '1. Base URL'}</Title>
            <Text>{apiBaseUrl}</Text>
            <div className='flex flex-wrap gap-3'>
              <Button theme='solid' type='primary' onClick={() => handleCopy(apiBaseUrl)}>
                {isChinese ? '复制 Base URL' : 'Copy Base URL'}
              </Button>
              <Link to='/console/token'>
                <Button>{isChinese ? '查看我的密钥' : 'Open Tokens'}</Button>
              </Link>
            </div>
            <Paragraph type='secondary'>
              {isChinese
                ? 'OpenAI 兼容接口建议直接使用这个 Base URL。常用端点包括 /chat/completions、/responses、/embeddings、/images/generations。'
                : 'Use this as the OpenAI-compatible base URL. Common endpoints include /chat/completions, /responses, /embeddings, and /images/generations.'}
            </Paragraph>
          </Space>
        </Card>

        <Card>
          <Space vertical spacing='medium' style={{ width: '100%' }}>
            <Title heading={4}>
              {isChinese ? '2. 鉴权方式' : '2. Authentication'}
            </Title>
            <pre className='overflow-x-auto rounded-xl bg-zinc-950 p-4 text-sm text-zinc-100'>
{`Authorization: Bearer YOUR_API_KEY`}
            </pre>
            <Paragraph type='secondary'>
              {isChinese
                ? '把 YOUR_API_KEY 替换成你控制台里看到的 sk- 开头密钥。不要把这个密钥暴露到前端网页或公开仓库。'
                : 'Replace YOUR_API_KEY with your sk- token from the console. Do not expose it in public frontend code or repositories.'}
            </Paragraph>
          </Space>
        </Card>

        <Card>
          <Space vertical spacing='medium' style={{ width: '100%' }}>
            <Title heading={4}>
              {isChinese ? '3. cURL 示例' : '3. cURL Example'}
            </Title>
            <pre className='overflow-x-auto rounded-xl bg-zinc-950 p-4 text-sm text-zinc-100'>
              {examples.curl}
            </pre>
            <Button onClick={() => handleCopy(examples.curl)}>
              {isChinese ? '复制 cURL' : 'Copy cURL'}
            </Button>
          </Space>
        </Card>

        <div className='grid gap-6 md:grid-cols-2'>
          <Card>
            <Space vertical spacing='medium' style={{ width: '100%' }}>
              <Title heading={4}>Python</Title>
              <pre className='overflow-x-auto rounded-xl bg-zinc-950 p-4 text-sm text-zinc-100'>
                {examples.python}
              </pre>
              <Button onClick={() => handleCopy(examples.python)}>
                {isChinese ? '复制 Python 示例' : 'Copy Python'}
              </Button>
            </Space>
          </Card>

          <Card>
            <Space vertical spacing='medium' style={{ width: '100%' }}>
              <Title heading={4}>JavaScript</Title>
              <pre className='overflow-x-auto rounded-xl bg-zinc-950 p-4 text-sm text-zinc-100'>
                {examples.javascript}
              </pre>
              <Button onClick={() => handleCopy(examples.javascript)}>
                {isChinese ? '复制 JavaScript 示例' : 'Copy JavaScript'}
              </Button>
            </Space>
          </Card>
        </div>

        <Card>
          <Space vertical spacing='medium' style={{ width: '100%' }}>
            <Title heading={4}>
              {isChinese ? '4. 模型怎么选' : '4. Choosing Models'}
            </Title>
            <Paragraph>
              {isChinese
                ? '可用模型名称以你站点控制台和模型广场中展示的为准。调用时把 model 字段替换成你账号可见的模型名，例如 gpt-4o-mini、claude-3-5-sonnet、gemini-2.0-flash。'
                : 'Use the model IDs shown in your console and model marketplace. Replace the model field with an available model such as gpt-4o-mini, claude-3-5-sonnet, or gemini-2.0-flash.'}
            </Paragraph>
            <Paragraph type='secondary'>
              {isChinese
                ? '如果某个模型不可用，通常是管理员还没有给该分组配置对应渠道。'
                : 'If a model is unavailable, the admin usually has not mapped a channel for your group yet.'}
            </Paragraph>
          </Space>
        </Card>
      </Space>
    </div>
  );
};

export default Docs;
