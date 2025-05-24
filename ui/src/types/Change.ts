export interface IChange {
  revision: string;
  timestamp: string;
  message: string;
  authorName?: string;
  authorEmail?: string;
  scmUrl?: string;
}
