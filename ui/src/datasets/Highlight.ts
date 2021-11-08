import {DatasetRow} from "./Dataset";

export class Highlight {
    alt = "";
    source = "";

    static fromDatasetRow(data: DatasetRow): Highlight {
        const highlight = new Highlight();
        highlight.alt = data.get("Alt") ?? "";
        highlight.source = data.get("Source") ?? "";
        return highlight;
    }
}
