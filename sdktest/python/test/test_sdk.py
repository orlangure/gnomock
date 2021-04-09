import gnomock
from gnomock.rest import ApiException
from urllib.parse import urlparse
import os

import unittest
import os

class TestSDK(unittest.TestCase):
    def setUp(self):
        addr = "127.0.0.1"

        dh = os.environ.get("DOCKER_HOST")
        if dh:
            u = urlparse(dh)
            addr = u.hostname

        host = "http://{}:23042".format(addr)

        configuration = gnomock.Configuration(host=host)
        with gnomock.ApiClient(configuration) as client:
            self.api = gnomock.PresetsApi(client)

    def tearDown(self):
        return super().tearDown()

    def test_mongo(self):
        options = gnomock.Options()
        file_name = os.path.abspath("./test/testdata/mongo")
        preset = gnomock.Mongo(data_path=file_name, version="3")
        mongo_request = gnomock.MongoRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_mongo(mongo_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_mysql(self):
        options = gnomock.Options()
        file_name = os.path.abspath("./test/testdata/mysql/schema.sql")
        preset = gnomock.Mysql(queries_files=[file_name], version="8")
        mysql_request = gnomock.MysqlRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_mysql(mysql_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_mariadb(self):
        options = gnomock.Options()
        file_name = os.path.abspath("./test/testdata/mysql/schema.sql")
        preset = gnomock.Mariadb(queries_files=[file_name], version="10")
        mariadb_request = gnomock.MariadbRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_mariadb(mariadb_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_mssql(self):
        options = gnomock.Options()
        file_name = os.path.abspath("./test/testdata/mssql/schema.sql")
        preset = gnomock.Mssql(queries_files=[file_name], license=True, version="2019-latest")
        mssql_request = gnomock.MssqlRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_mssql(mssql_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_postgres(self):
        options = gnomock.Options()
        file_name = os.path.abspath("./test/testdata/postgres/schema.sql")
        preset = gnomock.Postgres(queries_files=[file_name], version="12")
        postgres_request = gnomock.PostgresRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_postgres(postgres_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_redis(self):
        options = gnomock.Options(debug=True)
        values = {
            "foo": "bar",
            "number": 42,
            "float": 3.14
        }
        preset = gnomock.Redis(version="5", values=values)
        redis_request = gnomock.RedisRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_redis(redis_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_memcached(self):
        options = gnomock.Options()
        values = {
            "foo": "bar",
        }
        byte_values = {
            "key": "Z25vbW9jawo="
        }
        preset = gnomock.Memcached(version="1.6.6-alpine", values=values,
                byte_values=byte_values)
        memcached_request = gnomock.MemcachedRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_memcached(memcached_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_splunk(self):
        options = gnomock.Options()
        file_name = os.path.abspath("./test/testdata/splunk/events.jsonl")
        preset = gnomock.Splunk(values_file=file_name, accept_license=True,
                admin_password="12345678", version="8.0.2.1")
        splunk_request = gnomock.SplunkRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_splunk(splunk_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_localstack(self):
        options = gnomock.Options()
        preset = gnomock.Localstack(services=['s3'], version="0.11.0")
        localstack_request = gnomock.LocalstackRequest(options=options, preset=preset)
        id = ""

        try:
            response = self.api.start_localstack(localstack_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)

    def test_rabbitmq(self):
        file_name = os.path.abspath("./test/testdata/rabbitmq/messages.jsonl")
        options = gnomock.Options()
        message = gnomock.RabbitmqMessage(queue="alerts",
                content_type="text/plain", string_body="python")
        preset = gnomock.Rabbitmq(version="3.8.5-alpine",
                messages_files=[file_name], messages=[message])
        rabbitmq_request = gnomock.RabbitmqRequest(options=options,
                preset=preset)
        id = ""

        try:
            response = self.api.start_rabbit_mq(rabbitmq_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_kafka(self):
        options = gnomock.Options()
        preset = gnomock.Kafka(version="2.5.1-L0")
        kafka_request = gnomock.KafkaRequest(options=options,
                preset=preset)
        id = ""

        try:
            response = self.api.start_kafka(kafka_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_elastic(self):
        options = gnomock.Options()
        preset = gnomock.Elastic(version="7.9.2")
        elastic_request = gnomock.ElasticRequest(options=options,
                preset=preset)
        id = ""

        try:
            response = self.api.start_elastic(elastic_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_kubernetes(self):
        options = gnomock.Options()
        preset = gnomock.Kubernetes(version="latest")
        kubernetes_request = gnomock.KubernetesRequest(options=options,
                preset=preset)
        id = ""

        try:
            response = self.api.start_kubernetes(kubernetes_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


    def test_cockroachdb(self):
        options = gnomock.Options()
        preset = gnomock.Cockroachdb(version="latest")
        cockroachdb_request = gnomock.CockroachdbRequest(options=options,
                preset=preset)
        id = ""

        try:
            response = self.api.start_cockroach_db(cockroachdb_request)
            id = response.id
            self.assertEqual("127.0.0.1", response.host)

        finally:
            if id != "":
                stop_request = gnomock.StopRequest(id=id)
                self.api.stop(stop_request)


### gnomock-generator


if __name__ == "__main__":
    unittest.main()
