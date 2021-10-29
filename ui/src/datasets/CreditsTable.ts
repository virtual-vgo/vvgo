import {Credit} from "./Credit";

export type CreditsTable = Array<CreditsTopic>

export interface CreditsTopic {
    Name: string;
    Rows: Array<CreditsTeamRow>;
}

export interface CreditsTeamRow {
    Name: string;
    Rows: Credit[];
}
