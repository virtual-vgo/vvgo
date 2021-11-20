import { CSSProperties, useEffect, useState } from "react";
import { Badge } from "react-bootstrap";
import { latestProject, useProjects } from "../../datasets";

const BandcampIframeMaxWidth = 700;

export function Footer() {
  const [hideOverlay, setHideOverlay] = useState(false);
  const overlayHeight = hideOverlay ? 0 : 42;
  const overlayMaxWidth =
    window.innerWidth <= 700 ? BandcampIframeMaxWidth : 500;

  return (
    <footer>
      <div
        style={{
          height: overlayHeight + 38,
          width: "100%",
          textAlign: "center",
        }}
      >
        <a
          className="text-light"
          href="https://www.youtube.com/channel/UCeipEtsfjAA_8ATsd7SNAaQ"
        >
          <i className="fab fa-youtube fa-2x" />
        </a>
        <a className="text-light" href="https://vvgo.bandcamp.com/">
          <i className="fab fa-bandcamp fa-2x" />
        </a>
        <a className="text-light" href="https://github.com/virtual-vgo/vvgo">
          <i className="fab fa-github fa-2x" />
        </a>
        <a className="text-light" href="https://www.instagram.com/virtualvgo/">
          <i className="fab fa-instagram fa-2x" />
        </a>
        <a className="text-light" href="https://twitter.com/virtualvgo">
          <i className="fab fa-twitter fa-2x" />
        </a>
        <a className="text-light" href="https://discord.gg/vvgo">
          <i className="fab fa-discord fa-2x" />
        </a>
      </div>
      <BandcampOverlay
        hideOverlay={hideOverlay}
        setHideOverlay={setHideOverlay}
        maxWidth={overlayMaxWidth}
        height={overlayHeight}
        size="small"
        bgcol="8c17d9"
        linkcol="9a64ff"
        tracklist={false}
        artwork="none"
        transparent={true}
      />
    </footer>
  );
}

const BandcampOverlay = (props: {
  hideOverlay: boolean;
  setHideOverlay: (val: boolean) => void;
  maxWidth: number;
  height: number;
  size: "small" | "large";
  bgcol: string;
  linkcol: string;
  tracklist: boolean;
  artwork: "none" | "small" | "big";
  transparent: boolean;
}) => {
  const maxWidth = Math.min(props.maxWidth, BandcampIframeMaxWidth);
  const [hideX, setHideX] = useState(true);
  const project = latestProject(useProjects());
  if (!project) return <div />;
  if (project.BandcampAlbum == "") return <div />;
  if (props.hideOverlay) return <div />;

  const src =
    `https://bandcamp.com/EmbeddedPlayer/` +
    [
      `album=${project.BandcampAlbum}`,
      `size=${props.size}`,
      `bgcol=${props.bgcol}`,
      `linkcol=${props.linkcol}`,
      `tracklist=${props.tracklist ?? false}`,
      `artwork=${props.artwork ?? "none"}`,
      `transparent=${props.transparent ?? true}`,
    ].join("/");
  return (
    <>
      <iframe
        style={{
          position: "fixed",
          bottom: 0,
          left: 0,
          width: "100%",
          maxWidth: maxWidth,
          height: props.height,
        }}
        src={src}
        seamless
        onLoad={() => {
          if (window.innerWidth > maxWidth) setHideX(false);
        }}
      />
      <Badge
        hidden={hideX}
        style={{
          position: "fixed",
          bottom: props.height - 15,
          left: maxWidth - 15,
          cursor: "pointer",
        }}
        onClick={() => props.setHideOverlay(true)}
      >
        x
      </Badge>
    </>
  );
};
