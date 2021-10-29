import React = require("react");
import {Footer} from "./Footer";
import {Navbar} from "./Navbar";

export const RootContainer = (props: { title?: string, children: JSX.Element | JSX.Element[] }) => {
    if (props.title && props.title.length > 0) document.title = "VVGO | " + props.title;

    return <div className={"container"}>
        <Navbar/>
        {props.children}
        <Footer/>
    </div>;
};
