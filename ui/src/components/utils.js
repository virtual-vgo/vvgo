import React from "react";
import {makeStyles} from "@material-ui/core";

export function YoutubeIframe(props) {
    const classes = makeStyles({
        projectIframeWrapper: {
            position: "relative",
            paddingTop: "56.25%",
            textAlign: "center",
        },
        projectIframe: {
            position: "absolute",
            top: 0,
            left: 0,
            width: "100%",
            height: "100%"
        }
    })()

    const width = (window.width * 9) / 10
    const height = (9 * width) / 16
    return <div className={classes.projectIframeWrapper}>
        <iframe className={classes.projectIframe} height={height} width={width} src={props.src}
                allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                title="latest.Title"
                allowFullScreen/>
    </div>;
}
