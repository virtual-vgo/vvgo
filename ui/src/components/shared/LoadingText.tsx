import { randElement } from "../../utils";

export const LoadingText = () => {
  const loadingText = randElement([
    "😩 【Ｌｏａｄｉｎｇ】 😩",
    "(っ◔◡◔)っ ♥ 𝐿𝑜𝒶𝒹𝒾𝓃𝑔 ♥",
    "𝒲𝑒'𝓁𝓁 𝒷𝑒 𝓇𝒾𝑔𝒽𝓉 𝓌𝒾𝓉𝒽 𝓎𝑜𝓊 😘",
    "😳👌  ⓛＯα𝓓𝕚ＮＧ  💗🍩",
    "🐏  🎀  𝒯𝐻𝐸 𝐸𝒜𝑅𝒯𝐻 𝐼𝒮 𝐹𝐿𝒜𝒯  🎀  🐏",
  ]);

  return <h1 className="text-center">{loadingText}</h1>;
};
