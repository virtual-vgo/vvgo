import React = require('react');
import _ = require("lodash");
import {Render} from "./render";
import {
    Project,
    projectIsOpenForSubmission,
    projectIsPostProduction,
    projectIsReleased,
    useParts,
    useProjects
} from "./datasets";
import {Container} from "./components";
import {
    Box,
    Grid,
    Paper,
    Tab,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Tabs,
    Typography
} from "@material-ui/core";

export const Entrypoint = (selectors: string) => Render(<Projects/>, selectors)

const Projects = () => {
    const projects = useProjects()
    const parts = useParts()
    projects.sort((a, b) => (a.Name < b.Name) ? 1 : -1)
    const openForSubmission = projects.filter(p => projectIsOpenForSubmission(p))
    const postProduction = projects.filter(p => projectIsPostProduction(p))
    const released = projects.filter(p => projectIsReleased(p))

    const [value, setValue] = React.useState(0)
    const handleChange = (event: any, newValue: React.SetStateAction<number>) => {
        setValue(newValue)
    }

    const selection = _.get(_.concat(openForSubmission, postProduction, released), value, new Project()) as Project
    const selectionParts = parts.filter(p => p.Project === selection.Name)
    selectionParts.sort((a, b) => a.ScoreOrder - b.ScoreOrder)
    const partsTable = <Table className={'text-light'}>
        <TableHead>
            <TableCell>Part</TableCell>
            <TableCell>Downloads</TableCell>
        </TableHead>
        <TableBody>
            {selectionParts.map(p =>
                <TableRow key={p.PartName}>
                    <TableCell>{p.PartName}</TableCell>
                    <TableCell>Downloads...</TableCell>
                </TableRow>)}
        </TableBody>
    </Table>

    const selectionBox = <Box>
        <Typography variant={'h2'}>{selection.Title}</Typography>
        <Typography variant={'h4'}>Downloads...</Typography>
        <Paper>{partsTable}</Paper>
    </Box>


    return <Container>
        <Box>
            <Typography variant="h3" gutterBottom>
                Projects
            </Typography>
        </Box>
        <Grid container>
            <Grid item>
                <Tabs
                    orientation="vertical"
                    value={value}
                    onChange={handleChange}
                >
                    {openForSubmission.map(p => <Tab key={p.Title} label={p.Title}/>)}
                    {postProduction.map(p => <Tab key={p.Title} label={p.Title}/>)}
                    {released.map(p => <Tab key={p.Title} label={p.Title}/>)}
                </Tabs>
            </Grid>
            <Grid item>
                {selectionBox}
            </Grid>
        </Grid>
    </Container>
}
