import * as _ from "lodash";

export function randElement<T>(arr: Array<T>): T {
    if (arr.length == 0) return {} as T;
    return arr[_.random(arr.length)];
}
