import apiClient from '@/lib/services/core/api-client';
import {ApiErrorBase} from '@/lib/services/core/errors';
import type {InternalAxiosRequestConfig} from 'axios';

export interface LegacyApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}

/**
 * OpenFlare legacy 业务 API 基类
 * 解析 { success, message, data } 响应格式
 */
export class LegacyOpenFlareBaseService {
  protected static readonly basePath: string = '';

  protected static getFullPath(path: string): string {
    const full = `${this.basePath}${path}`;
    // Avoid Next.js 308 (strip slash) <-> Gin 301 (add slash) redirect loops in dev proxy.
    if (full.length > 1 && full.endsWith('/')) {
      return full.slice(0, -1);
    }
    return full;
  }

  protected static parseLegacyResponse<T>(body: LegacyApiResponse<T>): T {
    if (!body.success) {
      throw new ApiErrorBase(body.message || '请求失败');
    }
    return body.data;
  }

  protected static async legacyGet<T>(
    path: string,
    params?: Record<string, unknown>,
    config?: InternalAxiosRequestConfig,
  ): Promise<T> {
    const response = await apiClient.get<LegacyApiResponse<T>>(
      this.getFullPath(path),
      { ...config, params } as InternalAxiosRequestConfig,
    );
    return this.parseLegacyResponse(response.data);
  }

  protected static async legacyPost<T>(
    path: string,
    data?: unknown,
    config?: InternalAxiosRequestConfig,
  ): Promise<T> {
    const response = await apiClient.post<LegacyApiResponse<T>>(
      this.getFullPath(path),
      data,
      config,
    );
    return this.parseLegacyResponse(response.data);
  }
}
