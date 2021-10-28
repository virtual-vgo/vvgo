import * as _ from "lodash";

export const sleep = (ms: number): Promise<NodeJS.Timeout> => new Promise(resolve => setTimeout(resolve, ms));

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
