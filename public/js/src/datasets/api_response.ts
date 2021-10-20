import {Part} from "./part";
import {Session} from "./session";
import {Project} from "./project";

const _ = require("lodash");

export const Endpoint = '/api/v1'

export const ApiResponseStatus = Object.freeze({
    OK: "ok",
    Error: "error",
})

export class ApiResponse {
    Status: string
    Error?: ErrorResponse
    Dataset?: Array<Object>
    Parts?: Part[]
    Projects?: Project[]
    Sessions?: Session[]
    Identity?: Session
}

export class ErrorResponse {
    Code: Number
    Error: string
}
