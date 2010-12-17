using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.IO;
using System.Xml;
using Microsoft.Build.Evaluation;

namespace MSBuild.XCode
{
    public static class XProjectMerge
    {
        private static void Merge(Dictionary<string, List<Element>> main, Dictionary<string, List<Element>> template)
        {
            foreach (KeyValuePair<string, List<Element>> template_group in template)
            {
                if (main.ContainsKey(template_group.Key))
                {
                    // Merge
                    List<Element> mainElementsList;
                    main.TryGetValue(template_group.Key, out mainElementsList);

                    Dictionary<string, Element> mainElementsDict = new Dictionary<string, Element>();
                    foreach (Element e in mainElementsList)
                    {
                        mainElementsDict.Add(e.Name, e);
                    }

                    foreach (Element e in template_group.Value)
                    {
                        if (mainElementsDict.ContainsKey(e.Name))
                        {
                            // Merge element if concatenation of the values is required
                            if (e.Concat)
                            {
                                Element this_e;
                                mainElementsDict.TryGetValue(e.Name, out this_e);
                                this_e.Value = this_e.Value + e.Separator + e.Value;
                            }
                        }
                        else
                        {
                            // Add element
                            mainElementsList.Add(e.Copy());
                        }
                    }
                }
                else
                {
                    // Clone
                    List<Element> elements = new List<Element>();
                    main.Add(template_group.Key, elements);
                    foreach (Element e in template_group.Value)
                        elements.Add(e.Copy());
                }
            }
        }

        private static void Merge(Platform main, Platform template)
        {
            Merge(main.groups, template.groups);

            foreach (KeyValuePair<string, Config> p in main.configs)
            {
                Merge(p.Value.groups, main.groups);

                Config x;
                if (template.configs.TryGetValue(p.Key, out x))
                {
                    Merge(p.Value.groups, x.groups);
                }
            }
        }

        public static void Merge(Project main, Project template)
        {
            Merge(main.groups, template.groups);
            
            foreach (KeyValuePair<string, Platform> p in main.Platforms)
            {
                Merge(p.Value.groups, main.groups);

                Platform x;
                if (template.Platforms.TryGetValue(p.Key, out x))
                {
                    Merge(p.Value, x);
                }
            }
        }
    }
}