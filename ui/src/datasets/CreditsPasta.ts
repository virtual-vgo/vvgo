import { get, isEmpty } from "lodash/fp";

export class CreditsPasta {
  websitePasta = "";
  videoPasta = "";
  youtubePasta = "";

  static fromApiObject(obj: object): CreditsPasta | undefined {
    if (isEmpty(obj)) return undefined;
    const pasta = new CreditsPasta();
    pasta.websitePasta = get("WebsitePasta", obj) ?? "";
    pasta.websitePasta = get("VideoPasta", obj) ?? "";
    pasta.youtubePasta = get("YoutubePasta", obj) ?? "";
    return pasta;
  }
}
