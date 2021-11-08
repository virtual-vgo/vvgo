import {DatasetRow} from "./Dataset";

export class Credit {
    bottomText: string = "";
    majorCategory: string = "";
    minorCategory: string = "";
    name: string = "";
    order: string = "";
    project: string = "";

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
