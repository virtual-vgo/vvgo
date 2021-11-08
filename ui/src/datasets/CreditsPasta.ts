import {get} from "lodash/fp";

export class CreditsPasta {
    websitePasta: string = "";
    videoPasta: string = "";
    youtubePasta: string = "";

    static fromApiJSON(obj: object): CreditsPasta {
        const pasta = new CreditsPasta();
        pasta.websitePasta = get("WebsitePasta", obj) ?? "";
        pasta.websitePasta = get("VideoPasta", obj) ?? "";
        pasta.youtubePasta = get("YoutubePasta", obj) ?? "";
        return pasta;
    }
}
