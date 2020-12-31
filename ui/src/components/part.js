import React from "react";
import {ButtonGroup, Typography} from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Container from "@material-ui/core/Container";
import IconButton from "@material-ui/core/IconButton";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import CardMedia from "@material-ui/core/CardMedia"
import PlayArrowIcon from '@material-ui/icons/PlayArrow';
import makeStyles from "@material-ui/core/styles/makeStyles";

export default function Part(props) {
    function SheetMusic() {
        if (props.part.SheetMusicLink !== "") {
            return <div>
                Sheet Music: <object style={{width: '100%', height: '90vh'}} data={props.part.SheetMusicLink}/>
            </div>
        } else {
            return null
        }
    }

    function ClickTrack() {
        if (props.part.ClickTrackLink !== "") {
            return <div>
                Click Track: <audio controls src={props.part.ClickTrackLink}/>
            </div>
        } else {
            return null
        }
    }

    function ReferenceTrack() {
        if (props.project.ReferenceTrack !== "") {
            return <div>
                Reference Track: <audio controls src={props.project.ReferenceTrack}/>
            </div>
        } else {
            return null
        }
    }

    props.setAppTitle(`${props.project.Title} | ${props.part.PartName}`)
    document.title = `${props.project.Title} | ${props.part.PartName}`
    console.log("displaying", props.part)
    return <Container>
        <ButtonGroup variant='outlined'>
            <ProjectLinks {...props.project}/>
            <PartDownloads {...props.part}/>
        </ButtonGroup>
        <ReferenceTrack/>
        <ClickTrack/>
        <SheetMusic/>
    </Container>
}

function ProjectInfo(props) {
    return <div>
        <Typography paragraph>
            {props.Composers}<br/>
            <small>{props.Arrangers}</small>
        </Typography>
        <Typography variant='h4'>
            <strong>Submission Deadline:</strong> <em>{props.SubmissionDeadline} (Hawaii Time)</em>
        </Typography>
    </div>
}

function ProjectLinks(props) {
    return [
        {
            href: "https://www.youtube.com/watch?v=VgqtZ30bMgM",
            startIcon: <i className="fab fa-youtube"/>,
            children: 'Recording Instructions'
        },
        {
            href: props.SubmissionLink,
            startIcon: <i className="fab fa-dropbox"/>,
            children: 'Submit Recordings'
        },
        {
            href: props.ReferenceTrack,
            startIcon: <i className="far fa-file-audio"/>,
            children: 'Reference Track'
        },
    ].map(button => <Button key={button.href} {...button}/>)
}

function PartDownloads(props) {
    return [
        {
            href: props.SheetMusicLink,
            startIcon: <i className="far fa-file-pdf"/>,
            children: 'sheet music'
        },
        {
            href: props.ClickTrackLink,
            startIcon: <i className="far fa-file-audio"/>,
            children: 'click track'
        },
        {
            href: props.ConductorVideo,
            startIcon: <i className="far fa-file-video"/>,
            children: 'conductor video'
        },
        {
            href: props.PronunciationGuideLink,
            startIcon: <i className="fas fa-language"/>,
            children: 'pronunciation guide'
        },
    ].filter(b => b.href !== "").map(button => <Button key={button.href} {...button}/>)
}

const useStyles = makeStyles((theme) => ({
    root: {
        display: 'flex',
    },
    details: {
        display: 'flex',
        flexDirection: 'column',
    },
    content: {
        flex: '1 0 auto',
    },
    cover: {
        width: 151,
    },
    controls: {
        display: 'flex',
        alignItems: 'center',
        paddingLeft: theme.spacing(1),
        paddingBottom: theme.spacing(1),
    },
    playIcon: {
        height: 38,
        width: 38,
    },
}));

function MediaControlCard(props) {
    const classes = useStyles();
    return <Card className={classes.root}>
        <div className={classes.details}>
            <CardContent className={classes.content}>
                <Typography component="h5" variant="h5">
                    {props.part.PartName} - Click Track
                </Typography>
                <Typography variant="subtitle1" color="textSecondary">
                    {props.project.Title}
                </Typography>
            </CardContent>
            <div className={classes.controls}>
                <IconButton aria-label="play/pause">
                    <PlayArrowIcon className={classes.playIcon}/>
                </IconButton>
            </div>
        </div>
        <CardMedia
            className={classes.cover}
            src={props.part.ClickTrackLink}
            title={`${props.part.PartName} - Click Track`}
        />
    </Card>
}
