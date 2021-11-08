import {toPairs} from "lodash/fp";

export type DatasetRow = Map<string, string>

export class Dataset extends Array<DatasetRow> {
    static fromApiJSON(objs: object[]): Dataset {
        return objs?.map(row =>
            toPairs(row).reduce((a, [k, v]) => a.set(k, v), new Map<string, string>())) ?? [];
    }
}
