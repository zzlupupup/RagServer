export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
}
