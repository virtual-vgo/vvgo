import React from 'react'
import {Button, ButtonGroup} from "@material-ui/core";
import {makeStyles} from "@material-ui/core/styles";

export default function Footer() {
    return <footer>
        <SocialMediaRow/>
        <PolicyRow/>
    </footer>
}

const useStyles = makeStyles((theme) => ({
    root: {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        '& > *': {
            margin: theme.spacing(1),
        },
    },
}));

function SocialMediaRow() {
    const classes = useStyles();
    return <div className={classes.root}>
        <ButtonGroup variant="text">
            <Button href="https://www.youtube.com/channel/UCeipEtsfjAA_8ATsd7SNAaQ">
                <i className="fab fa-youtube fa-2x"/>
            </Button>
            <Button href="https://www.facebook.com/groups/1080154885682377/">
                <i className="fab fa-facebook fa-2x"/>
            </Button>
            <Button href="https://vvgo.bandcamp.com/">
                <i className="fab fa-bandcamp fa-2x"/>
            </Button>
            <Button href="https://github.com/virtual-vgo/vvgo">
                <i className="fab fa-github fa-2x"/>
            </Button>
            <Button href="https://www.instagram.com/virtualvgo/">
                <i className="fab fa-instagram fa-2x"/>
            </Button>
            <Button href="https://twitter.com/virtualvgo">
                <i className="fab fa-twitter fa-2x"/>
            </Button>
            <Button href="https://discord.com/invite/9RVUJMQ">
                <i className="fab fa-discord fa-2x"/>
            </Button>
        </ButtonGroup>
    </div>
}

function PolicyRow() {
    const classes = useStyles();
    return <div className={classes.root}>
        <ButtonGroup size="small" variant="text">
            <Button href="https://vvgo.org/privacy">privacy policy</Button>
            <Button href="https://vvgo.org/cookie-policy">cookie policy</Button>
        </ButtonGroup>
    </div>
}
