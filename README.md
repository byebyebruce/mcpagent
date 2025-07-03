# MCP Agent

## 编译
```bash
go build .
```
## 准备
1. 设置OpenAI环境变量 或者 cp `.env.exmple` to `.env` 修改里面配置
2. 修改mcp server 配置 cp `mcp.json.example' to `mcp.json` 修改里面配置

## MCP 资源
- [ModelScop MCP](https://www.modelscope.cn/mcp)
- [mcpmarket](https://mcpmarket.com/zh)

## 用法
1. 交互式
```bash
./mcpagent
```

2. 输入提示词
```bash
./mcpagent --i '获取https://www.gov.cn/guoqing/2021-05/13/content_5606149.htm 页面内容,获取里面的民族人口数据,地区人口,年龄结构,性别构成等数据,然后画成饼图'
Assistant:
Tool: web.read_page Arguments: {"url":"https://www.gov.cn/guoqing/2021-05/13/content_5606149.htm"}
Calling web.read_page
Tool: mcp-server-chart.generate_pie_chart Arguments: {"title": "民族人口构成", "data": [{"category": "汉族", "value": 1286311334}, {"category": "少数民族", "value": 125467390}]}
Tool: mcp-server-chart.generate_pie_chart Arguments: {"title": "地区人口构成", "data": [{"category": "东部地区", "value": 563717119}, {"category": "中部地区", "value": 364694362}, {"category": "西部地区", "value": 382852295}, {"category": "东北地区", "value": 98514948}]}
Tool: mcp-server-chart.generate_pie_chart Arguments: {"title": "人口年龄结构", "data": [{"category": "0—14岁", "value": 253383938}, {"category": "15—59岁", "value": 894376020}, {"category": "60岁及以上", "value": 264018766}]}
Tool: mcp-server-chart.generate_pie_chart Arguments: {"title": "人口性别构成", "data": [{"category": "男性", "value": 723339956}, {"category": "女性", "value": 688438768}]}
Calling mcp-server-chart.generate_pie_chart
Calling mcp-server-chart.generate_pie_chart
Calling mcp-server-chart.generate_pie_chart
Calling mcp-server-chart.generate_pie_chart
已为您生成四个不同主题的人口数据饼图：

1. [民族人口构成](https://mdn.alipayobjects.com/one_clip/afts/img/JdWFSJVBK1kAAAAAQvAAAAgAoEACAQFr/original)
2. [地区人口构成](https://mdn.alipayobjects.com/one_clip/afts/img/XVq3Qbv6tbAAAAAARFAAAAgAoEACAQFr/original)
3. [人口年龄结构](https://mdn.alipayobjects.com/one_clip/afts/img/NC2vTYRCrU4AAAAAQ_AAAAgAoEACAQFr/original)
4. [人口性别构成](https://mdn.alipayobjects.com/one_clip/afts/img/VyKLRLLiX9AAAAAAQqAAAAgAoEACAQFr/original)

您可以点击链接查看各个饼图的具体信息。             
```
 - 例子里用到的画图mcp server [ntvis/mcp-server-chart](https://www.modelscope.cn/mcp/servers/@antvis/mcp-server-chart)
3. 启动agent sse mcp server
```bash
./mcpagent server http://localhost:8080
```

