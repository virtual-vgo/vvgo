import * as _ from "lodash";

export function randElement<T>(arr: Array<T>): T {
    return arr[_.random(arr.length)]
}
