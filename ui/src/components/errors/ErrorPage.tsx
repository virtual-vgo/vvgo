import React = require("react");
import {Link} from "react-router-dom";

export const ErrorPage = (props: { src: string, alt: string }) => {
    document.title = props.alt;
    return <div className="mt-2">
        <img src={props.src}
             alt={props.alt}
             style={{
                 margin: "auto",
                 maxHeight: "100vh",
                 maxWidth: "100vh",
                 width: "100%",
                 height: "auto",
                 display: "block",
                 borderRadius: "5px",
                 borderColor: "#9600de",
                 borderWidth: "2px",
                 borderStyle: " solid",
             }}/>
        <div className="d-flex justify-content-center">
            <Link to="/" style={{fontWeight: "bold", color: "ghostwhite"}}>
                Click here to return to safety.
            </Link>
        </div>
    </div>;
};
