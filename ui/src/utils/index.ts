import _ from "lodash";

export function randElement<T>(arr: Array<T>): T | undefined {
    return _.isEmpty(arr) ? undefined : arr[_.random(arr.length - 1)];
}
