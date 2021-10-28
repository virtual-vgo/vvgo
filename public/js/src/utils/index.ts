import * as _ from "lodash";

export function randElement<T>(arr: Array<T>): T {
    switch (true) {
        case !arr:
            return null as T;
        case arr.length == 0:
            return {} as T;
        default:
            return arr[_.random(arr.length - 1)];
    }
}
