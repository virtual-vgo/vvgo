import React from 'react'
import Helmet from "react-helmet";
import {Container, Divider, TableBody, Typography} from "@material-ui/core";
import Link from "@material-ui/core/Link";
import Table from "@material-ui/core/Table";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";

export default function About(props) {
    return <Container>
        <Helmet>
            <title>About</title>
            <meta name="description" content="About VVGO"/>
        </Helmet>
        <Typography variant="h2">About VVGO</Typography>
        <Divider/>
        <Info/>
        <LeaderTable leaders={props.leaders}/>
    </Container>
}

function Info() {
    const typographyProps = {paragraph: true, align: "justify"}
    return <div style={{marginTop: '15px'}}>
        <Typography {...typographyProps}>
            Formed in March 2020, <strong>Virtual Video Game Orchestra</strong> (VVGO, "vee-vee-go") is an online
            volunteer-run music ensemble predicated on providing a musical performance outlet for musicians whose
            IRL rehearsals and performances were cancelled due to COVID-19. Led and organized by members from
            various video game ensembles, and with a community of hundreds of musicians from across the globe,
            VVGO is open to any who wish to participate regardless of instrument, skill level, or musical background.
        </Typography>
        <Typography {...typographyProps}>
            Our mission is to provide a fun and accessible virtual community of musicians from around the world
            through performing video game music.
        </Typography>
        <Typography {...typographyProps}>
            We are always accepting new members into our community. If you would like to join our orchestra or
            get more information about our current performance opportunities, please join us on
            <Link href="https://discord.gg/9RVUJMQ">Discord</Link>!
        </Typography>
    </div>
}

function LeaderTable(props) {
    return <div>
        <Typography variant="h3">VVGO Leadership</Typography>
        <Table>
            <TableBody>
                {props.leaders.map(leader => <LeaderRow key={leader.Name} {...leader}/>)}
            </TableBody>
        </Table>
    </div>
}

function LeaderRow(props) {
    function LeaderName() {
        if (props.Email !== "") {
            let href = "mailto: " + props.Email
            return <Link href={href} color='inherit'>
                {props.Name}<br/><small>{props.Epithet}</small>
            </Link>
        } else {
            return <Typography>
                {props.Name}<br/><small>{props.Epithet}</small>
            </Typography>
        }
    }

    return <TableRow>
        <TableCell><img src={props.Icon} alt={props.Name} height="125"/></TableCell>
        <TableCell><LeaderName/></TableCell>
        <TableCell>
            <Typography paragraph>{props.Blurb}</Typography>
            <Typography paragraph><i>{props.Affiliations}</i></Typography>
        </TableCell>
    </TableRow>
}
