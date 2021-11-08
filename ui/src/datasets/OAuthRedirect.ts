import {get} from "lodash/fp";

export class OAuthRedirect {
    DiscordURL: string = "";
    State: string = "";
    Secret: string = "";

    static fromApiJSON(obj: object): OAuthRedirect {
        const data = new OAuthRedirect();
        data.DiscordURL = get("DiscordURL", obj);
        data.State = get("State", obj);
        data.Secret = get("Secret", obj);
        return data;
    }
}
