import { Credit } from "./Credit";

export type CreditsTable = CreditsTopic[];

export interface CreditsTopic {
  Name: string;
  Rows: Array<CreditsTeamRow>;
}

export interface CreditsTeamRow {
  Name: string;
  Rows: Credit[];
}
