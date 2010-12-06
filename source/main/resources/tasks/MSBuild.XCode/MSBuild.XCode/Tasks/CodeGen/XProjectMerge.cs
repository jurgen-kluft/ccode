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
        private static void Merge(Dictionary<string, List<XElement>> main, Dictionary<string, List<XElement>> template)
        {
            foreach (KeyValuePair<string, List<XElement>> template_group in template)
            {
                if (main.ContainsKey(template_group.Key))
                {
                    // Merge
                    List<XElement> mainElementsList;
                    main.TryGetValue(template_group.Key, out mainElementsList);

                    Dictionary<string, XElement> mainElementsDict = new Dictionary<string, XElement>();
                    foreach (XElement e in mainElementsList)
                    {
                        mainElementsDict.Add(e.Name, e);
                    }

                    foreach (XElement e in template_group.Value)
                    {
                        if (mainElementsDict.ContainsKey(e.Name))
                        {
                            // Merge element if concatenation of the values is required
                            if (e.Concat)
                            {
                                XElement this_e;
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
                    List<XElement> elements = new List<XElement>();
                    main.Add(template_group.Key, elements);
                    foreach (XElement e in template_group.Value)
                        elements.Add(e.Copy());
                }
            }
        }

        private static void Merge(XPlatform main, XPlatform template)
        {
            Merge(main.groups, template.groups);

            foreach (KeyValuePair<string, XConfig> p in main.configs)
            {
                Merge(p.Value.groups, main.groups);

                XConfig x;
                if (template.configs.TryGetValue(p.Key, out x))
                {
                    Merge(p.Value.groups, x.groups);
                }
            }
        }

        public static void Merge(XProject main, XProject template)
        {
            Merge(main.groups, template.groups);
            
            foreach (KeyValuePair<string, XPlatform> p in main.platforms)
            {
                Merge(p.Value.groups, main.groups);

                XPlatform x;
                if (template.platforms.TryGetValue(p.Key, out x))
                {
                    Merge(p.Value, x);
                }
            }
        }
    }
}