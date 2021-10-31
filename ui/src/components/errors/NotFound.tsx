import React = require("react");
import imgSrc = require("./404.gif");
import {ErrorPage} from "./ErrorPage";

export const NotFound = () => <ErrorPage src={imgSrc} alt="404 Not Found"/>;