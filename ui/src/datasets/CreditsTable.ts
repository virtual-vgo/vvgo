import { get } from "lodash/fp";
import { Credit } from "./Credit";

export class CreditsTeamRow {
  Name: string;
  Rows: Credit[];

  constructor(name?: string, rows?: Credit[]) {
    this.Name = name ?? "";
    this.Rows = rows ?? [];
  }

  static fromApiObject(obj: object): CreditsTeamRow {
    const rows = get("Rows", obj)?.map((r: object) => Credit.fromApiObject(r));
    return new CreditsTeamRow(get("Name", obj), rows);
  }
}

export class CreditsTopic {
  Name: string;
  Rows: CreditsTeamRow[];

  constructor(name?: string, rows?: CreditsTeamRow[]) {
    this.Name = name ?? "";
    this.Rows = rows ?? [];
  }

  static fromApiObject(obj: object): CreditsTopic {
    const rows = get("Rows", obj)?.map((r: object) =>
      CreditsTeamRow.fromApiObject(r)
    );
    return new CreditsTopic(get("Name", obj), rows);
  }
}

export class CreditsTable extends Array<CreditsTopic> {
  static fromApiArray(objs: object[]): CreditsTable | undefined {
    return objs?.map((r) => CreditsTopic.fromApiObject(r));
  }
}
