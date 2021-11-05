import * as _ from "lodash";

export function randElement<T>(arr: Array<T>): T {
    return _.isEmpty(arr) ? null as unknown as T : arr[_.random(arr.length - 1)];
}
