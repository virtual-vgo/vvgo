import React from 'react';
import clsx from 'clsx';
import {makeStyles, useTheme} from '@material-ui/core/styles';
import Drawer from '@material-ui/core/Drawer';
import CssBaseline from '@material-ui/core/CssBaseline';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import {Link as RouterLink} from "react-router-dom";

const drawerWidth = 240;

const useStyles = makeStyles((theme) => ({
    root: {display: 'flex'},
    appBar: {
        transition: theme.transitions.create(['margin', 'width'], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
    },
    appBarShift: {
        width: `calc(100% - ${drawerWidth}px)`,
        marginLeft: drawerWidth,
        transition: theme.transitions.create(['margin', 'width'], {
            easing: theme.transitions.easing.easeOut,
            duration: theme.transitions.duration.enteringScreen,
        }),
    },
    menuButton: {marginRight: theme.spacing(2)},
    hide: {display: 'none'},
    drawer: {width: drawerWidth, flexShrink: 0},
    drawerPaper: {width: drawerWidth},
    drawerHeader: {
        display: 'flex',
        alignItems: 'center',
        padding: theme.spacing(0, 1),
        // necessary for content to be below app bar
        ...theme.mixins.toolbar,
        justifyContent: 'flex-end',
    },
    content: {
        flexGrow: 1,
        padding: theme.spacing(3),
        transition: theme.transitions.create('margin', {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
        }),
        marginLeft: -drawerWidth,
    },
    contentShift: {
        transition: theme.transitions.create('margin', {
            easing: theme.transitions.easing.easeOut,
            duration: theme.transitions.duration.enteringScreen,
        }),
        marginLeft: 0,
    },
}));

export default function AppDrawer(props) {
    const classes = useStyles();
    const theme = useTheme();
    const [open, setOpen] = React.useState(false);

    const openDrawer = () => setOpen(true);
    const closeDrawer = () => setOpen(false);

    return (
        <div className={classes.root}>
            <CssBaseline/>
            <AppBar position="fixed" className={clsx(classes.appBar, {[classes.appBarShift]: open})}>
                <Toolbar>
                    <IconButton aria-label="open drawer" color="inherit" edge="start" onClick={openDrawer}
                                className={clsx(classes.menuButton, open && classes.hide)}>
                        <MenuIcon/>
                    </IconButton>
                    <Typography variant="h6" noWrap>Virtual VGO</Typography>
                </Toolbar>
            </AppBar>
            <Drawer className={classes.drawer} variant="persistent" anchor="left"
                    open={open} classes={{paper: classes.drawerPaper,}}>
                <div className={classes.drawerHeader}>
                    <IconButton onClick={closeDrawer}>
                        {theme.direction === 'ltr' ? <ChevronLeftIcon/> : <ChevronRightIcon/>}
                    </IconButton>
                </div>
                <Divider/>
                <List>
                    <MyListItem to={"/about"}>About</MyListItem>
                    <MyListItem to={"/releases"}>Releases</MyListItem>
                    <MemberListItem roles={props.uiRoles.data} to={"/parts"}>Parts</MemberListItem>
                </List>
            </Drawer>
            <main className={clsx(classes.content, {[classes.contentShift]: open})}>
                <div className={classes.drawerHeader}/>
                {props.children}
            </main>
        </div>
    );
}

function MyListItem(props) {
    return <ListItem button component={RouterLink} to={props.to}>
        <ListItemText primary={props.children}/>
    </ListItem>
}

function MemberListItem(props) {
    if (props.roles.includes("vvgo-member")) {
        return <MyListItem to={props.to}>{props.children}</MyListItem>
    } else {
        return null
    }
}

function TeamsListItem(props) {
    if (props.roles.includes("vvgo-teams")) {
        return <ListItem button component={RouterLink} to={props.to} color='secondary'>
            <ListItemText primary={props.children}/>
        </ListItem>
    } else {
        return null
    }
}
