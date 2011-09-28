using System;
using System.Collections;
using System.Collections.Generic;
using MySql.Data.MySqlClient;
using MSBuild.XCode.Helpers;
using MSBuild.XCode;

namespace xpackage_repo
{
    class PackageDatabaseMySQL : IPackageDatabase
    {
        private MySqlConnection sqlConnect;

        #region IPackageDatabase Members

        public bool connect(string databaseURL)
        {
            try
            {
                // Example:
                // "server=cnshasap2;port=3307;database=gamedev;uid=developer;password=Fastercode189";
                //string connectionString = "server=cnshasap2;port=3307;database=gamedev;uid=developer;password=Fastercode189";
                string connectionString = databaseURL;
                sqlConnect = new MySqlConnection(connectionString);
                sqlConnect.Open();
                return true;
            }
            catch (Exception)
            {
                sqlConnect = null;
                return false;
            }
        }

        public void disconnect()
        {
            if (sqlConnect != null)
            {
                sqlConnect.Close();
                sqlConnect = null;
            }
        }

        private MySqlCommand sqlCreateCmd(string Sqlstr)
        {
            MySqlCommand Cmd = new MySqlCommand();
            Cmd.Connection = sqlConnect;
            Cmd.CommandText = Sqlstr;
            return Cmd;
        }

        private int sqlExecuteNonQuery(string Sqlstr)
        {
            try
            {
                MySqlCommand Cmd = sqlCreateCmd(Sqlstr);
                int intResult = Cmd.ExecuteNonQuery();
                return intResult;
            }
            catch
            {
                return 0;
            }
        }

        private object sqlRunExecuteScalar(string Sqlstr)
        {
            MySqlCommand Cmd = sqlCreateCmd(Sqlstr);
            object obResult = Cmd.ExecuteScalar();
            return obResult;
        }

        private MySqlDataReader sqlRunExecuteReader(string Sqlstr)
        {
            MySqlCommand Cmd = sqlCreateCmd(Sqlstr);
            MySqlDataReader Result = Cmd.ExecuteReader();
            return Result;
        }

        private ArrayList sqlReadFromTable(string Sqlstr)
        {
            ArrayList result = new ArrayList();
            MySqlDataReader reader = sqlRunExecuteReader(Sqlstr);
            while (reader.Read())
            {
                for (int i = 0; i < reader.FieldCount; ++i)
                    result.Add(reader[i]);
            }
            reader.Close();
            return result;
        }

        private int insertPackage(string package_name, string package_group, string package_language)
        {
            string group_search_Sql = string.Format("select idPackageGroup from packagegroup_pg where Name_pg = '{0}'", package_group);
            ArrayList idPackageGroups = sqlReadFromTable(group_search_Sql);
            if (idPackageGroups.Count == 0)
            {
                string insert_group_sql = string.Format("insert into packagegroup_pg(Name_pg) values('{0}')", package_group);
                if (sqlExecuteNonQuery(insert_group_sql) == 0)
                    return 0;
                idPackageGroups = sqlReadFromTable(group_search_Sql);
            }
            int idPackageGroup = (int)idPackageGroups[0];

            string lang_search_Sql = string.Format("select idPackageLanguage from packagelanguage_pl where Name_pl = '{0}'", package_language);
            ArrayList idPackageLanguages = sqlReadFromTable(lang_search_Sql);
            if (idPackageLanguages.Count == 0)
            {
                string insert_language_sql = string.Format("insert into packagelanguage_pl(Name_pl) values('{0}')", package_language);
                if (sqlExecuteNonQuery(insert_language_sql) == 0)
                    return 0;
                idPackageLanguages = sqlReadFromTable(lang_search_Sql);
            }
            int idPackageLanguage = (int)idPackageLanguages[0];

            string package_search_Sql = string.Format("select idPackage_pk from package_pk where Name_pk = '{0}' AND idPackageGroup_pk = {1} AND idPackageLanguage_pk = {2}", package_name, idPackageGroup, idPackageLanguage);
            ArrayList idPackages = sqlReadFromTable(package_search_Sql);
            if (idPackages.Count == 0)
            {
                string insert_package_sql = string.Format("insert into package_pk(idPackageGroup_pk, idPackageLanguage_pk, Name_pk) values({0},{1},'{2}')", idPackageGroup, idPackageLanguage, package_name);
                if (sqlExecuteNonQuery(insert_package_sql) == 0)
                    return 0;
                idPackages = sqlReadFromTable(package_search_Sql);
            }
            int idPackage = (int)idPackages[0];
            return idPackage;
        }

        private int findPackage(string package_name, string package_group, string package_language)
        {
            string group_search_Sql = string.Format("select idPackageGroup from packagegroup_pg where Name_pg = '{0}'", package_group);
            ArrayList idPackageGroups = sqlReadFromTable(group_search_Sql);
            if (idPackageGroups.Count == 0)
                return 0;
            int idPackageGroup = (int)idPackageGroups[0];

            string lang_search_Sql = string.Format("select idPackageLanguage from packagelanguage_pl where Name_pl = '{0}'", package_language);
            ArrayList idPackageLanguages = sqlReadFromTable(lang_search_Sql);
            if (idPackageLanguages.Count == 0)
                return 0;
            int idPackageLanguage = (int)idPackageLanguages[0];

            string package_search_Sql = string.Format("select idPackage_pk from package_pk where Name_pk = '{0}' AND idPackageGroup_pk = {1} AND idPackageLanguage_pk = {2}", package_name, idPackageGroup, idPackageLanguage);
            ArrayList idPackages = sqlReadFromTable(package_search_Sql);
            if (idPackages.Count == 0)
                return 0;
            int idPackage = (int)idPackages[0];
            return idPackage;
        }

        private int insertPlatform(string platform)
        {
            string search_Sql = string.Format("select idPackagePlatform from packageplatform_pp where Name_pp = '{0}'", platform);
            ArrayList ids = sqlReadFromTable(search_Sql);
            if (ids.Count == 0)
            {
                string insert_sql = string.Format("insert into packageplatform_pp(Name_pp) values('{0}')", platform);
                if (sqlExecuteNonQuery(insert_sql) == 0)
                    return 0;
                ids = sqlReadFromTable(search_Sql);
            }
            int id = (int)(ids[0]);
            return id;
        }

        private int insertBranch(string branch)
        {
            string search_Sql = string.Format("select idPackageBranch from packagebranch_pb where Name_pb = '{0}'", branch);
            ArrayList ids = sqlReadFromTable(search_Sql);
            if (ids.Count == 0)
            {
                string insert_sql = string.Format("insert into packagebranch_pb(Name_pb) values('{0}')", branch);
                if (sqlExecuteNonQuery(insert_sql) == 0)
                    return 0;
                ids = sqlReadFromTable(search_Sql);
            }
            int id = (int)ids[0];
            return id;
        }

        private int findPlatform(string platform)
        {
            string search_Sql = string.Format("select idPackagePlatform from packageplatform_pp where Name_pp = '{0}'", platform);
            ArrayList ids = sqlReadFromTable(search_Sql);
            if (ids.Count == 0)
                    return 0;
            int id = (int)ids[0];
            return id;
        }

        private int findBranch(string branch)
        {
            string search_Sql = string.Format("select idPackageBranch from packagebranch_pb where Name_pb = '{0}'", branch);
            ArrayList ids = sqlReadFromTable(search_Sql);
            if (ids.Count == 0)
                    return 0;
            int id = (int)ids[0];
            return id;
        }

        private bool hasVersion(int idPackage, int idPackagePlatform, int idPackageBranch, PackageVersion_pv package)
        {
            string search_Sql = string.Format("select idPackageVersion_pv from packageversion_pv where idPackage_pv = {0} AND idPackagePlatform_pv = {1} AND idPackageBranch_pv = {2} AND Version_pv = {3}", idPackage, idPackagePlatform, idPackageBranch, package.Version);
            ArrayList idPackageVersions = sqlReadFromTable(search_Sql);
            return (idPackageVersions.Count > 0);
        }

        private int insertVersion(int idPackage, int idPackagePlatform, int idPackageBranch, PackageVersion_pv package)
        {
            string search_Sql = string.Format("select idPackageVersion_pv from packageversion_pv where idPackage_pv = {0} AND idPackagePlatform_pv = {1} AND idPackageBranch_pv = {2} AND Version_pv = {3}", idPackage, idPackagePlatform, idPackageBranch, package.Version);
            ArrayList idPackageVersions = sqlReadFromTable(search_Sql);
            if (idPackageVersions.Count == 0)
            {
                string insert_sql = string.Format("insert into packageversion_pv(idPackage_pv, idPackagePlatform_pv, idPackageBranch_pv, Date_pv, Version_pv, Location_pv, Changeset_pv) values({0},{1},{2},'{3}',{4},'{5}','{6}')", idPackage, idPackagePlatform, idPackageBranch, package.Datetime, package.Version, package.Location, package.Changeset);
                if (sqlExecuteNonQuery(insert_sql) == 0)
                    return 0;

                idPackageVersions = sqlReadFromTable(search_Sql);
                if (idPackageVersions.Count == 0)
                    return 0;
                int idPackageVersion = (int)idPackageVersions[0];
                return idPackageVersion;
            }
            else
            {
                int idPackageVersion = (int)idPackageVersions[0];
                return idPackageVersion;
            }
        }

        private int findVersion(PackageVersion_pv package)
        {
            int idPackage = findPackage(package.Name, package.Group, package.Language);
            if (idPackage == 0)
                return 0;

            int idPackagePlatform = findPlatform(package.Platform);
            if (idPackagePlatform == 0)
                return 0;

            int idPackageBranch = findBranch(package.Branch);
            if (idPackageBranch == 0)
                return 0;

            string packageversion_search_Sql = string.Format("select idPackageVersion_pv from packageversion_pv where idPackage_pv = {0} AND idPackagePlatform_pv = {1} and idPackageBranch_pv = {2} AND Version_pv = {3}", idPackage, idPackagePlatform, idPackageBranch, package.Version);
            ArrayList idPackageVersions = sqlReadFromTable(packageversion_search_Sql);
            if (idPackageVersions.Count == 0)
                return 0;

            int idPackageVersion = (int)idPackageVersions[0];
            return idPackageVersion;
        }

        private int findVersion(string pk_name, string pk_group, string pk_lang, string pk_platform, string pk_branch, Int64 pk_version)
        {
            int idPackage = findPackage(pk_name, pk_group, pk_lang);
            if (idPackage == 0)
                return 0;

            int idPackagePlatform = findPlatform(pk_platform);
            if (idPackagePlatform == 0)
                return 0;

            int idPackageBranch = findBranch(pk_branch);
            if (idPackageBranch == 0)
                return 0;

            string packageversion_search_Sql = string.Format("select idPackageVersion_pv from packageversion_pv where idPackage_pv = {0} AND idPackagePlatform_pv = {1} and idPackageBranch_pv = {2} AND Version_pv = {3}", idPackage, idPackagePlatform, idPackageBranch, pk_version);
            ArrayList idPackageVersions = sqlReadFromTable(packageversion_search_Sql);
            if (idPackageVersions.Count == 0)
                return 0;

            int idPackageVersion = (int)idPackageVersions[0];
            return idPackageVersion;
        }

        private int findLatestVersion(PackageVersion_pv package, Int64 start, bool include_start, Int64 end, bool include_end)
        {
            int idPackage = findPackage(package.Name, package.Group, package.Language);
            if (idPackage == 0)
                return 0;

            int idPackagePlatform = findPlatform(package.Platform);
            if (idPackagePlatform == 0)
                return 0;

            int idPackageBranch = findBranch(package.Branch);
            if (idPackageBranch == 0)
                return 0;

            string packageversion_search_Sql = string.Format("select idPackageVersion_pv from packageversion_pv where idPackage_pv = {0} AND idPackagePlatform_pv = {1} and idPackageBranch_pv = {2}", idPackage, idPackagePlatform, idPackageBranch);

            if (include_start)
                packageversion_search_Sql += string.Format(" AND Version_pv>={0}", start);
            else
                packageversion_search_Sql += string.Format(" AND Version_pv>{0}", start);

            if (include_end)
                packageversion_search_Sql += string.Format(" AND Version_pv<={0}", end);
            else
                packageversion_search_Sql += string.Format(" AND Version_pv<{0}", end);

            // Order (ascending), so latest version is at the end of the array
            packageversion_search_Sql += " ORDER BY Version_pv";

            ArrayList idPackageVersions = sqlReadFromTable(packageversion_search_Sql);
            if (idPackageVersions.Count == 0)
                return 0;

            // Take
            int idPackageVersion = (int)idPackageVersions[idPackageVersions.Count - 1];
            return idPackageVersion;        
        }

        public bool submit(PackageVersion_pv package, List<KeyValuePair<string, Int64>> dependencies)
        {
            int idPackage = insertPackage(package.Name, package.Group, package.Language);
            if (idPackage == 0)
            {
                Loggy.Error(String.Format("Error: Remote Repo Submit: failed to insert package"));
                return false;
            }

            int idPackagePlatform = insertPlatform(package.Platform);
            if (idPackagePlatform == 0)
            {
                Loggy.Error(String.Format("Error: Remote Repo Submit: failed to insert platform"));
                return false;
            }

            int idPackageBranch = insertBranch(package.Branch);
            if (idPackageBranch == 0)
            {
                Loggy.Error(String.Format("Error: Remote Repo Submit: failed to insert branch"));
                return false;
            }

            if (hasVersion(idPackage, idPackagePlatform, idPackageBranch, package))
            {
                // It seems we already have submitted this version
                return true;
            }
            else
            {
                int idPackageVersion = insertVersion(idPackage, idPackagePlatform, idPackageBranch, package);
                if (idPackageVersion == 0)
                {
                    Loggy.Error(String.Format("Error: Remote Repo Submit: failed to insert version"));
                    return false;
                }

                // Insert dependencies
                foreach (KeyValuePair<string, Int64> pv in dependencies)
                {
                    int idPackageDependency = findVersion(pv.Key, package.Group, package.Language, package.Platform, package.Branch, pv.Value);
                    if (idPackageDependency == 0)
                    {
                        ComparableVersion cv = new ComparableVersion(pv.Value);
                        Loggy.Error(String.Format("Error: Remote Repo Submit: dependency package {0} with version {1} hasn't been deployed yet", pv.Key, cv.ToString()));
                        return false;
                    }

                    string table_name = "packagedependency_pd";
                    string insert_sql = string.Format("insert into {0} values({1}, {2})", table_name, idPackageVersion, idPackageDependency);
                    if (sqlExecuteNonQuery(insert_sql) == 0)
                    {
                        Loggy.Error(String.Format("Error: Remote Repo Submit: failed to insert into table {0} using statement {1}", table_name, insert_sql));
                        return false;
                    }
                }
            }

            return true;
        }

        public bool findUniqueVersion(PackageVersion_pv package)
        {
            int idPackageVersion = findVersion(package);
            return idPackageVersion != 0;
        }

        public bool findLatestVersion(PackageVersion_pv package, out Int64 outVersion)
        {
            Int64 lowestVersion = PackageVersion_pv.LowestVersion;
            Int64 highestVersion = PackageVersion_pv.HighestVersion;

            int idPackageVersion = findLatestVersion(package, lowestVersion, true, highestVersion, true);
            if (idPackageVersion != 0)
            {
                string version_get_Sql = string.Format("select Version_pv from packageversion_pv where idPackageVersion_pv = {0}", idPackageVersion);
                outVersion = (int)sqlRunExecuteScalar(version_get_Sql);
            }
            else
            {
                outVersion = 0;
            }
            return outVersion != 0;
        }

        public bool findLatestVersion(PackageVersion_pv package, Int64 start_version, bool include_start, Int64 end_version, bool include_end, out Int64 outVersion)
        {
            int idPackageVersion = findLatestVersion(package, start_version, include_start, end_version, include_end);
            if (idPackageVersion != 0)
            {
                string version_get_Sql = string.Format("select Version_pv from packageversion_pv where idPackageVersion_pv = {0}", idPackageVersion);
                outVersion = (Int64)sqlRunExecuteScalar(version_get_Sql);
            }
            else
            {
                outVersion = 0;
            }
            return outVersion != 0;
        }

        public bool retrieveVarsOf(PackageVersion_pv package, out Dictionary<string, object> vars)
        {
            vars = new Dictionary<string, object>();

            int idPackageVersion = findVersion(package);
            if (idPackageVersion == 0)
                return false;

            string location_search_Sql = string.Format("select Date_pv, Location_pv, Changeset_pv from packageversion_pv where idPackageVersion_pv = {0}", idPackageVersion);
            ArrayList columns = sqlReadFromTable(location_search_Sql);
            if (columns.Count == 0)
                return false;

            // Take
            vars.Add("DateTime", (DateTime)columns[0]);
            vars.Add("Location", (string)columns[1]);
            vars.Add("ChangeSet", (string)columns[2]);

            return true;
        }

        #endregion
    }
}
