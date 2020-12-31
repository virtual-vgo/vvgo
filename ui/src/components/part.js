import React, {useState} from "react";
import Helmet from "react-helmet";
import {ButtonGroup, Typography} from "@material-ui/core";
import {Document, Page, pdfjs} from "react-pdf";
import ReactPlayer from "react-player";
import Button from "@material-ui/core/Button";

//https://github.com/wojtekmaj/react-pdf#standard-browserify-and-others
pdfjs.GlobalWorkerOptions.workerSrc = `//cdnjs.cloudflare.com/ajax/libs/pdf.js/${pdfjs.version}/pdf.worker.min.js`;

export default function Part(props) {
    function SheetMusic() {
        const [numPages, setNumPages] = useState(null);

        function onDocumentLoadSuccess({numPages}) {
            setNumPages(numPages);
        }



        if (props.part.SheetMusicLink !== "") {
            return <Document file={props.part.SheetMusicLink} onLoadSuccess={onDocumentLoadSuccess}>
                {Array.from(new Array(numPages), (el, index) => (
                    <Page renderMode='canvas' key={`page_${index + 1}`} pageNumber={index + 1}/>
                ))}
            </Document>
        } else {
            return null
        }
    }

    console.log("displaying", props.part)
    return <div>
        <Helmet>
            <title>{props.project.Title} | {props.part.PartName}</title>
            <meta name="description" content=""/>
        </Helmet>
        <Typography variant='h3'>{props.project.Title} - {props.part.PartName}</Typography>
        <ProjectInfo {...props.project}/>
        <ButtonGroup variant='outlined'>
            <ProjectLinks {...props.project}/>
            <PartDownloads {...props.part}/>
        </ButtonGroup>
        <SheetMusic/>
    </div>
}

function MediaPlayer(props) {
    if (ReactPlayer.canPlay(props.ClickTrackLink)) {
        console.log("can play 😀", props.Project, props.PartName, props.ClickTrackLink)
        return <ReactPlayer url={props.ClickTrackLink}/>
    } else {
        console.log("cant play yet 😩", props.Project, props.PartName)
        return null
    }
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
            href: props.ReferenceTrack,
            startIcon: <i className="far fa-file-audio"/>,
            children: 'Reference Track'
        },
        {
            href: props.SubmissionLink,
            startIcon: <i className="fab fa-dropbox"/>,
            children: 'Submit Recordings'
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
