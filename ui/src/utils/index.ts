import {isEmpty, random} from "lodash/fp";

export function randElement<T>(arr: Array<T>): T | undefined {
    return isEmpty(arr) ? undefined : arr[random(0, arr.length - 1)];
}
