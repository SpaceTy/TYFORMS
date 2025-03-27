using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace TYBLOG.Migrations
{
    /// <inheritdoc />
    public partial class AddAcceptanceStatus : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string>(
                name: "AcceptanceStatus",
                table: "Applications",
                type: "TEXT",
                nullable: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "AcceptanceStatus",
                table: "Applications");
        }
    }
}
