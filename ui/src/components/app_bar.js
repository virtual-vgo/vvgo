import React from "react";
import clsx from "clsx";
import {AppBar as MuiAppBar, IconButton, Toolbar, Typography} from "@material-ui/core";
import {ChevronRight} from '@material-ui/icons'
import {useVVGOStyles} from "./styles";

const useStyles = useVVGOStyles

export default function VVGOAppBar(props) {
    const classes = useStyles();
    return <MuiAppBar position="fixed"
                      className={clsx(classes.appBar, {[classes.appBarShift]: props.drawerState.isOpen})}>
        <Toolbar>
            <IconButton aria-label="open drawer" color="inherit" edge="start" onClick={props.drawerState.openDrawer}
                        className={clsx(classes.menuButton, props.drawerState.isOpen && classes.hide)}>
                <ChevronRight/>
            </IconButton>
            <Typography variant="h6" noWrap>{props.title}</Typography>
        </Toolbar>
    </MuiAppBar>
}
