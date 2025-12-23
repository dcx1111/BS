/**
 * MCP API客户端
 * 提供对话式图片检索的API接口
 */
import api from './client'

/**
 * 对话式图片搜索请求参数
 */
export interface MCPSearchRequest {
  query: string      // 自然语言查询字符串
  page?: number      // 页码，默认为1
  pageSize?: number  // 每页数量，默认为20
}

/**
 * 对话式图片搜索响应
 */
export interface MCPSearchResponse {
  query: string                    // 原始查询
  filters: Record<string, string>  // 转换后的过滤器
  total: number                    // 总数量
  page: number                     // 当前页码
  pageSize: number                 // 每页数量
  items: any[]                     // 图片列表
}

/**
 * 对话式图片搜索
 * 使用自然语言查询图片，AI会自动将查询转换为搜索条件
 * 
 * @param request 搜索请求参数
 * @returns 搜索结果
 */
export const mcpSearch = async (request: MCPSearchRequest): Promise<MCPSearchResponse> => {
  const response = await api.post<MCPSearchResponse>('/mcp/search', request)
  return response.data
}

