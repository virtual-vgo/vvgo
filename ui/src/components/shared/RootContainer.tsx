import _ = require("lodash");
import {Footer} from "./Footer";
import {Navbar} from "./Navbar";

export const RootContainer = (props: { title?: string, children: JSX.Element | JSX.Element[] }) => {
    document.title = _.isEmpty(props.title) ?
        "VVGO | Virtual Video Game Orchestra" :
        "VVGO | " + props.title;

    return <div className={"container"}>
        <Navbar/>
        {props.children}
        <Footer/>
    </div>;
};
