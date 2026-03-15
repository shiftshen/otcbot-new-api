import React, { useContext, useMemo } from 'react';
import { Button, Card, Space, Typography, Tag } from '@douyinfe/semi-ui';
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
  const currentModels = [
    'gpt-4o-mini',
    'gpt-4.1',
    'claude-3-5-sonnet',
    'gemini-2.0-flash',
  ];

  const examples = useMemo(
    () => ({
      curl: `curl ${apiBaseUrl}/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "用一句话介绍 OTCBot"}
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
      php: `<?php
require 'vendor/autoload.php';

$client = OpenAI::factory()
    ->withApiKey('YOUR_API_KEY')
    ->withBaseUri('${apiBaseUrl}')
    ->make();

$response = $client->chat()->create([
    'model' => 'gpt-4o-mini',
    'messages' => [
        ['role' => 'user', 'content' => 'Hello']
    ],
]);

echo $response->choices[0]->message->content;`,
      java: `OpenAIClient client = OpenAIOkHttpClient.builder()
    .apiKey("YOUR_API_KEY")
    .baseUrl("${apiBaseUrl}")
    .build();

ChatCompletionCreateParams params = ChatCompletionCreateParams.builder()
    .model("gpt-4o-mini")
    .addUserMessage("Hello")
    .build();

ChatCompletion completion = client.chat().completions().create(params);
System.out.println(completion.choices().get(0).message().content().get(0));`,
      responses: `curl ${apiBaseUrl}/responses \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "gpt-4.1",
    "input": "帮我写一句欢迎语"
  }'`,
      embeddings: `curl ${apiBaseUrl}/embeddings \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "text-embedding-3-small",
    "input": "OTCBot"
  }'`,
    }),
    [apiBaseUrl],
  );

  const handleCopy = async (value) => {
    const ok = await copy(value);
    if (ok) showSuccess(isChinese ? '已复制到剪贴板' : 'Copied');
  };

  const steps = isChinese
    ? [
        '注册并登录账号',
        '充值余额或购买套餐',
        '进入控制台创建 API Key',
        '把 Base URL 设置为本站 /v1',
        '在请求头里带上 Bearer 密钥',
      ]
    : [
        'Create and sign in to your account',
        'Recharge balance or purchase a subscription',
        'Create an API key in the console',
        'Set the Base URL to this site /v1',
        'Send the key as a Bearer token',
      ];

  const commonErrors = isChinese
    ? [
        {
          title: '401 Unauthorized',
          desc: '密钥不对、已删除，或者没有放在 Authorization: Bearer ... 里。',
        },
        {
          title: '余额不足 / quota not enough',
          desc: '账户余额不足，或者当前分组没有可用额度，需要先充值。',
        },
        {
          title: 'model not found',
          desc: '你请求的模型当前账号不可用，请改成控制台/模型广场里可见的模型名。',
        },
        {
          title: '429 Too Many Requests',
          desc: '当前模型或账号触发了限流，稍后重试或更换模型。',
        },
      ]
    : [
        {
          title: '401 Unauthorized',
          desc: 'The key is invalid, deleted, or not sent as Authorization: Bearer ...',
        },
        {
          title: 'quota not enough',
          desc: 'Your balance is insufficient or the current group has no usable quota.',
        },
        {
          title: 'model not found',
          desc: 'The requested model is not available to your account. Use one shown in the console.',
        },
        {
          title: '429 Too Many Requests',
          desc: 'Rate limits were hit for this account or model. Retry later or switch models.',
        },
      ];

  return (
    <div className='mx-auto max-w-6xl px-4 py-24'>
      <Space vertical spacing='loose' style={{ width: '100%' }}>
        <div className='max-w-3xl'>
          <Title heading={2}>
            {isChinese ? `${systemName} 接入文档` : `${systemName} Integration Guide`}
          </Title>
          <Paragraph type='secondary'>
            {isChinese
              ? '你充值后创建的 sk- 密钥，就是调用接口时放在 Authorization 里的 Bearer Token。本站兼容 OpenAI 风格接口。'
              : 'Your generated sk- key is the Bearer token used in Authorization. This site exposes OpenAI-compatible APIs.'}
          </Paragraph>
          <div className='flex flex-wrap gap-3'>
            <Link to='/console/token'>
              <Button theme='solid' type='primary'>
                {isChinese ? '去创建密钥' : 'Create API Key'}
              </Button>
            </Link>
            <Link to='/console/topup'>
              <Button>{isChinese ? '去充值' : 'Recharge'}
              </Button>
            </Link>
          </div>
        </div>

        <Card>
          <Space vertical spacing='medium' style={{ width: '100%' }}>
            <Title heading={4}>{isChinese ? '1. 先做什么' : '1. Quick Setup'}</Title>
            <div className='grid gap-3 md:grid-cols-5'>
              {steps.map((step, index) => (
                <div key={step} className='rounded-2xl border border-semi-color-border bg-semi-color-fill-0 p-4'>
                  <Text strong>{index + 1}</Text>
                  <Paragraph style={{ margin: '8px 0 0' }}>{step}</Paragraph>
                </div>
              ))}
            </div>
          </Space>
        </Card>

        <div className='grid gap-6 md:grid-cols-2'>
          <Card>
            <Space vertical spacing='medium' style={{ width: '100%' }}>
              <Title heading={4}>{isChinese ? '2. 调用地址' : '2. Base URL'}</Title>
              <Text>{apiBaseUrl}</Text>
              <div className='flex flex-wrap gap-3'>
                <Button theme='solid' type='primary' onClick={() => handleCopy(apiBaseUrl)}>
                  {isChinese ? '复制 Base URL' : 'Copy Base URL'}
                </Button>
              </div>
              <Paragraph type='secondary'>
                {isChinese
                  ? 'OpenAI 兼容接口建议统一使用这个 Base URL。常用端点包括 /chat/completions、/responses、/embeddings、/images/generations。'
                  : 'Use this as the OpenAI-compatible base URL. Common endpoints include /chat/completions, /responses, /embeddings, and /images/generations.'}
              </Paragraph>
            </Space>
          </Card>

          <Card>
            <Space vertical spacing='medium' style={{ width: '100%' }}>
              <Title heading={4}>{isChinese ? '3. 鉴权方式' : '3. Authentication'}</Title>
              <pre className='overflow-x-auto rounded-xl bg-zinc-950 p-4 text-sm text-zinc-100'>
{`Authorization: Bearer YOUR_API_KEY`}
              </pre>
              <Paragraph type='secondary'>
                {isChinese
                  ? '把 YOUR_API_KEY 替换成控制台里看到的 sk- 开头密钥。这个密钥等同于密码，不要公开。'
                  : 'Replace YOUR_API_KEY with the sk- key from the console. Treat it like a password and never expose it publicly.'}
              </Paragraph>
            </Space>
          </Card>
        </div>

        <Card>
          <Space vertical spacing='medium' style={{ width: '100%' }}>
            <Title heading={4}>{isChinese ? '4. 常用模型' : '4. Common Models'}</Title>
            <div className='flex flex-wrap gap-2'>
              {currentModels.map((model) => (
                <Tag key={model} size='large' color='blue'>
                  {model}
                </Tag>
              ))}
            </div>
            <Paragraph type='secondary'>
              {isChinese
                ? '最终以你账号在控制台和模型广场里可见的模型为准。模型名必须原样写入请求的 model 字段。'
                : 'Always use models visible to your account in the console and model marketplace. The model value must match exactly.'}
            </Paragraph>
          </Space>
        </Card>

        <Card>
          <Space vertical spacing='medium' style={{ width: '100%' }}>
            <Title heading={4}>{isChinese ? '5. cURL 示例' : '5. cURL Example'}</Title>
            <pre className='overflow-x-auto rounded-xl bg-zinc-950 p-4 text-sm text-zinc-100'>
              {examples.curl}
            </pre>
            <div className='flex flex-wrap gap-3'>
              <Button onClick={() => handleCopy(examples.curl)}>
                {isChinese ? '复制 cURL' : 'Copy cURL'}
              </Button>
              <Button onClick={() => handleCopy(examples.responses)}>
                {isChinese ? '复制 Responses 示例' : 'Copy Responses'}
              </Button>
              <Button onClick={() => handleCopy(examples.embeddings)}>
                {isChinese ? '复制 Embeddings 示例' : 'Copy Embeddings'}
              </Button>
            </div>
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

          <Card>
            <Space vertical spacing='medium' style={{ width: '100%' }}>
              <Title heading={4}>PHP</Title>
              <pre className='overflow-x-auto rounded-xl bg-zinc-950 p-4 text-sm text-zinc-100'>
                {examples.php}
              </pre>
              <Button onClick={() => handleCopy(examples.php)}>
                {isChinese ? '复制 PHP 示例' : 'Copy PHP'}
              </Button>
            </Space>
          </Card>

          <Card>
            <Space vertical spacing='medium' style={{ width: '100%' }}>
              <Title heading={4}>Java</Title>
              <pre className='overflow-x-auto rounded-xl bg-zinc-950 p-4 text-sm text-zinc-100'>
                {examples.java}
              </pre>
              <Button onClick={() => handleCopy(examples.java)}>
                {isChinese ? '复制 Java 示例' : 'Copy Java'}
              </Button>
            </Space>
          </Card>
        </div>

        <div className='grid gap-6 md:grid-cols-2'>
          <Card>
            <Space vertical spacing='medium' style={{ width: '100%' }}>
              <Title heading={4}>{isChinese ? '6. 余额和计费' : '6. Billing and Balance'}</Title>
              <Paragraph>
                {isChinese
                  ? '你的调用费用从账户余额或订阅额度中扣除。是否优先消耗订阅额度，取决于你在充值页看到的 Billing Preference 配置。'
                  : 'Usage is deducted from your account balance or subscription quota. Whether subscription quota is consumed first depends on the billing preference configured on your account.'}
              </Paragraph>
              <Paragraph type='secondary'>
                {isChinese
                  ? '如果接口返回余额不足，请先到充值页补余额，或确认当前账号所在分组是否有可用渠道。'
                  : 'If you receive insufficient balance errors, recharge first or confirm that your account group has active channels.'}
              </Paragraph>
            </Space>
          </Card>

          <Card>
            <Space vertical spacing='medium' style={{ width: '100%' }}>
              <Title heading={4}>{isChinese ? '7. 常见错误' : '7. Common Errors'}</Title>
              <Space vertical spacing='medium' style={{ width: '100%' }}>
                {commonErrors.map((item) => (
                  <div key={item.title} className='rounded-2xl border border-semi-color-border bg-semi-color-fill-0 p-4'>
                    <Text strong>{item.title}</Text>
                    <Paragraph style={{ margin: '6px 0 0' }} type='secondary'>
                      {item.desc}
                    </Paragraph>
                  </div>
                ))}
              </Space>
            </Space>
          </Card>
        </div>
      </Space>
    </div>
  );
};

export default Docs;
