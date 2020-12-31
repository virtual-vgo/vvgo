import {AppBar as MuiAppBar} from "@material-ui/core";
import clsx from "clsx";
import Toolbar from "@material-ui/core/Toolbar";
import IconButton from "@material-ui/core/IconButton";
import MenuIcon from "@material-ui/icons/Menu";
import Typography from "@material-ui/core/Typography";
import React from "react";
import {useVVGOStyles} from "./styles";

const useStyles = useVVGOStyles

export default function VVGOAppBar(props) {
    const classes = useStyles();
    return <MuiAppBar position="fixed"
                      className={clsx(classes.appBar, {[classes.appBarShift]: props.drawerState.isOpen})}>
        <Toolbar>
            <IconButton aria-label="open drawer" color="inherit" edge="start" onClick={props.drawerState.openDrawer}
                        className={clsx(classes.menuButton, props.drawerState.isOpen && classes.hide)}>
                <MenuIcon/>
            </IconButton>
            <Typography variant="h6" noWrap>{props.title}</Typography>
        </Toolbar>
    </MuiAppBar>
}
