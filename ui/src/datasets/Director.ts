import {DatasetRow} from "./Dataset";

export class Director {
    name = "";
    epithet = "";
    affiliations = "";
    blurb = "";
    icon = "";

    static fromDatasetRow(data: DatasetRow): Director {
        const director = new Director();
        director.name = data.get("Name") ?? "";
        director.epithet = data.get("Epithet") ?? "";
        director.affiliations = data.get("Affiliations") ?? "";
        director.blurb = data.get("Blurb") ?? "";
        director.icon = data.get("Icon") ?? "";
        return director;
    }
}
