export interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
}

export interface PageMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface PaginatedResponse<T> {
  items: T[];
  meta: PageMeta;
}

export interface PageParams {
  page?: number;
  limit?: number;
}
