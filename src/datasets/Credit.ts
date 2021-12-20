import { get } from "lodash/fp";
import { DatasetRow } from "./Dataset";

export class Credit {
  bottomText = "";
  majorCategory = "";
  minorCategory = "";
  name = "";
  order = "";
  project = "";

  static fromApiObject(obj: object): Credit {
    const credit = new Credit();
    credit.bottomText = get("BottomText", obj) ?? "";
    credit.majorCategory = get("MajorCategory", obj) ?? "";
    credit.minorCategory = get("MinorCategory", obj) ?? "";
    credit.name = get("Name", obj) ?? "";
    credit.order = get("Order", obj) ?? "";
    credit.project = get("Project", obj) ?? "";
    return credit;
  }

  static fromDatasetRow(row: DatasetRow): Credit {
    const credit = new Credit();
    credit.bottomText = row.get("BottomText") ?? "";
    credit.majorCategory = row.get("MajorCategory") ?? "";
    credit.minorCategory = row.get("MinorCategory") ?? "";
    credit.name = row.get("Name") ?? "";
    credit.order = row.get("Order") ?? "";
    credit.project = row.get("Project") ?? "";
    return credit;
  }
}
