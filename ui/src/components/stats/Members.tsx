import d3 = require("d3");
import _ = require("lodash");
import {VegaLite} from "react-vega";
import {useCredits} from "../../datasets";
import {RootContainer} from "../shared/RootContainer";

export const Members = () => {
    const credits = _.defaultTo(useCredits(), []).filter(x => !_.isEmpty(x.Name));
    if (_.isEmpty(credits)) return <div/>;

    const topCredited = d3.flatRollup(credits, g => g.length, p => p.Name)
        .map(([name, count]) => ({name, count}))
        .sort((a, b) => a.count - b.count)
        .reverse()
        .slice(0, 10);

    const performers = credits.filter(x => x.MajorCategory != "CREW");
    const topCreditedPerformers = d3.flatRollup(performers, g => g.length, p => p.Name)
        .map(([name, count]) => ({name, count}))
        .sort((a, b) => a.count - b.count)
        .reverse()
        .slice(0, 10);

    const topUniqueParts = d3.flatRollup(performers, g => new Set(g).size, p => p.Name)
        .map(([name, count]) => ({name, count}))
        .sort((a, b) => a.count - b.count)
        .reverse()
        .slice(0, 10);

    return <RootContainer title={"Member Stats"}>
        <div className={"d-grid justify-content-center"}>
            <h1>Member Stats</h1>
            <VegaLite
                data={{data: topCredited}}
                spec={{
                    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
                    title: "Top Credited Overall",
                    background: "transparent",
                    data: {name: "data"},
                    config: {title: {color: "white", fontSize: 30}},
                    height: 600,
                    width: 800,
                    mark: {type: "bar", tooltip: true},
                    encoding: {
                        y: {
                            field: "name",
                            type: "nominal",
                            axis: {title: "", labelColor: "white", labelFontSize: 15},
                            sort: "-x",
                        },
                        x: {
                            field: "count",
                            type: "quantitative",
                            axis: {
                                title: "Submissions",
                                titleColor: "white",
                                labelColor: "white",
                                labelFontSize: 15,
                            },
                        },
                    },
                }}/>

            <VegaLite
                data={{data: topCreditedPerformers}}
                spec={{
                    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
                    title: "Top Credited Performers",
                    background: "transparent",
                    data: {name: "data"},
                    config: {title: {color: "white", fontSize: 30}},
                    height: 600,
                    width: 800,
                    mark: {type: "bar", tooltip: true},
                    encoding: {
                        y: {
                            field: "name",
                            type: "nominal",
                            axis: {title: "", labelColor: "white", labelFontSize: 15},
                            sort: "-x",
                        },
                        x: {
                            field: "count",
                            type: "quantitative",
                            axis: {
                                title: "Submissions",
                                titleColor: "white",
                                labelColor: "white",
                                labelFontSize: 15,
                            },
                        },
                    },
                }}/>

            <VegaLite
                data={{data: topUniqueParts}}
                spec={{
                    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
                    title: "Most Unique Credited",
                    data: {name: "data"},
                    background: "transparent",
                    config: {title: {color: "white", fontSize: 30}},
                    height: 600,
                    width: 800,
                    mark: {type: "bar", tooltip: true},
                    encoding: {
                        y: {
                            field: "name",
                            type: "nominal",
                            axis: {title: "", labelColor: "white", labelFontSize: 15},
                            sort: "-x",
                        },
                        x: {
                            field: "count",
                            type: "quantitative",
                            axis: {
                                title: "Submissions",
                                titleColor: "white",
                                labelColor: "white",
                                labelFontSize: 15,
                            },
                        },
                    },
                }}/>
        </div>
    </RootContainer>;
};
