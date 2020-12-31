import React from "react";

const styles = {
    projectIframeWrapper: {
        position: "relative",
        paddingTop: "56.25%",
        textAlign: "center",
    },
    projectIframe: {
        position: "absolute",
        top: "0",
        left: "0",
        width: "100%",
        height: "100%"
    }
}

export function YoutubeIframe(props) {
    const width = (window.width * 9) / 10
    const height = (9 * width) / 16
    return <div className={styles.projectIframeWrapper}>
        <iframe className={styles.projectIframe} height={height} width={width} src={props.src}
                allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                title="latest.Title"
                allowFullScreen/>
    </div>;
}
