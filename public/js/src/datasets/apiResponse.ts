import {Part} from "./part";
import {Session} from "./session";
import {Project} from "./project";
import {GuildMember} from "./guildMember";

const _ = require("lodash");

export const Endpoint = '/api/v1'

export const ApiStatuses = Object.freeze({
    OK: "ok",
    Error: "error",
})

export type ApiStatus = "ok" | "error"

export class ApiResponse {
    Status: string
    Error?: ErrorResponse
    Dataset?: Array<Object>
    Parts?: Part[]
    Projects?: Project[]
    Sessions?: Session[]
    Identity?: Session
    GuildMembers?: GuildMember[]
}

export class ErrorResponse {
    Code: Number
    Error: string
}
