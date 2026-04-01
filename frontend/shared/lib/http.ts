import axios from "axios";

import type { ApiResponse } from "@/shared/types/api";

export const http = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1",
  headers: {
    "Content-Type": "application/json",
  },
});

http.interceptors.response.use(
  (res) => res,
  (err) => {
    const message: string =
      err.response?.data?.error ?? err.message ?? "Unexpected error";
    return Promise.reject(new Error(message));
  },
);

export async function apiGet<T>(
  url: string,
  params?: Record<string, unknown>,
): Promise<T> {
  const res = await http.get<ApiResponse<T>>(url, { params });
  return res.data.data;
}

export async function apiPost<T>(url: string, body?: unknown): Promise<T> {
  const res = await http.post<ApiResponse<T>>(url, body);
  return res.data.data;
}

export async function apiPatch<T>(url: string, body?: unknown): Promise<T> {
  const res = await http.patch<ApiResponse<T>>(url, body);
  return res.data.data;
}
