default_run_options[:pty] = true
ssh_options[:forward_agent] = true


set :application, "SiteMonitor"
set :repository,  "git@github.com:curt-labs/SiteMonitor.git"

set :scm, :git
set :scm_passphrase, ""
set :user, "deployer"

#role :web, "curt-api-server1.cloudapp.net", "curt-api-server2.cloudapp.net"
role :web, "173.255.117.20", "173.255.112.170"

set :deploy_to, "/home/#{user}/#{application}"
set :deploy_via, :remote_cache

set :use_sudo, false
set :sudo_prompt, ""
set :normalize_asset_timestamps, false

set :default_environment, {
  'GOPATH' => "$HOME/gocode"
}

after :deploy, "deploy:goget", "db:configure", "deploy:compile", "deploy:stop", "deploy:restart"

namespace :db do
  desc "set database Connection String"
  task :configure do
    set(:database_username) { Capistrano::CLI.ui.ask("Database Username:") }
  
    set(:database_password) { Capistrano::CLI.password_prompt("Database Password:") }

    db_config = <<-EOF
      package database

      const (
        db_proto = "tcp"
        db_addr  = "curtsql.cloudapp.net:3306"
        db_user  = "#{database_username}"
        db_pass  = "#{database_password}"
        db_name  = "SiteMonitor"
      )
    EOF
    run "mkdir -p #{deploy_to}/current/helpers/database"
    put db_config, "#{deploy_to}/current/helpers/database/ConnectionString.go"
  end
end
namespace :deploy do
  task :goget do
  	run "/home/#{user}/bin/go get github.com/ziutek/mymysql/native"
  	run "/home/#{user}/bin/go get github.com/ziutek/mymysql/mysql"
  end
  task :compile do
  	run "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 /home/#{user}/bin/go build -o #{deploy_to}/current/site-monitor #{deploy_to}/current/index.go"
  end
  task :start do ; end
  task :stop do 
      kill_processes_matching "site-monitor"
  end
  task :restart do
  	restart_cmd = "#{current_release}/site-monitor -http=127.0.0.1:8090 -path=#{deploy_to}/current/"
  	run "nohup sh -c '#{restart_cmd} &' > #{application}-nohup.out"
  end
end

def kill_processes_matching(name)
  begin
    run "killall -q #{name}"
  rescue Exception => error
    puts "No processes to kill"
  end
end
